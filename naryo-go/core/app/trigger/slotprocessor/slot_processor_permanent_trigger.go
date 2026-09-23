package slotprocessor

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/dispatch"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/interactor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/logging"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/trigger"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/block"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/common"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

// SlotProcessorPermanentTrigger is the PermanentTrigger that, for every
// SlotEvent it processes, asks the BlockInteractor for the block at that
// slot and dispatches the resulting SolanaBlockEvent back into the
// Dispatcher. It ignores every other event.Event.
//
// A slot the leader skipped (block.ErrSlotSkipped) will never resolve to a
// block, so it moves on to the next slot rather than treating it as a
// failure. Every other unresolved slot — one the cluster hasn't reached yet,
// or whose block BlockInteractor reports as not yet available
// (block.ErrBlockNotAvailable) — is retried with retry's backoff policy,
// since it must eventually be processed. Exhausting the retry budget is not
// expected to happen in normal operation; when it does, Process returns an
// error and leaves the decision of whether to retry to its caller.
type SlotProcessorPermanentTrigger struct {
	blockInteractor interactor.BlockInteractor
	dispatcher      dispatch.Dispatcher
	retry           *common.RetryConfiguration
}

// NewSlotProcessorPermanentTrigger builds a SlotProcessorPermanentTrigger
// backed by blockInteractor and dispatcher, retrying an unresolved slot per
// retry's backoff policy.
func NewSlotProcessorPermanentTrigger(
	blockInteractor interactor.BlockInteractor,
	dispatcher dispatch.Dispatcher,
	retry *common.RetryConfiguration,
) *SlotProcessorPermanentTrigger {
	return &SlotProcessorPermanentTrigger{
		blockInteractor: blockInteractor,
		dispatcher:      dispatcher,
		retry:           retry,
	}
}

// Process asks the BlockInteractor for the block at e's slot and dispatches
// the resulting SolanaBlockEvent. Process is only ever called with a
// SlotEvent, since Supports rejects every other event.Event; a returned
// error is logged and isolated by the Dispatcher that called Process.
//
// A skipped slot is not an error: it is logged at debug level and Process
// returns nil without dispatching anything.
func (t *SlotProcessorPermanentTrigger) Process(ctx context.Context, e event.Event) error {
	slotEvent, ok := e.(event.SlotEvent)
	if !ok {
		return nil
	}

	b, err := t.getBlock(ctx, slotEvent.Slot)
	if errors.Is(err, block.ErrSlotSkipped) {
		logging.DebugWithCtx(ctx, "slot processor trigger: slot skipped, no block to dispatch", "slot", slotEvent.Slot)
		return nil
	}
	if err != nil {
		return fmt.Errorf("slot processor trigger: failed to get block for slot %d: %w", slotEvent.Slot, err)
	}

	blockEvent, err := toSolanaBlockEvent(slotEvent.NodeID(), b)
	if err != nil {
		return fmt.Errorf("slot processor trigger: failed to build block event for slot %d: %w", slotEvent.Slot, err)
	}

	t.dispatcher.Dispatch(ctx, blockEvent)
	return nil
}

// getBlock fetches the block at slot. Before every attempt, it checks the
// node's current slot (BlockInteractor.GetSlot): while slot is still ahead
// of it, the cluster hasn't reached it yet, so it waits — per t.retry's
// backoff policy — without even calling GetBlock. Once slot is at or behind
// the current slot, GetBlock is called; a block.ErrBlockNotAvailable result
// is retried the same way, since the node may simply not have caught up
// with its own reported slot yet. A skipped slot, or any other error
// (including a failed GetSlot check), is returned immediately without
// retrying. If the retry budget is exhausted, the last error observed is
// returned.
func (t *SlotProcessorPermanentTrigger) getBlock(ctx context.Context, slot uint64) (*block.SolanaBlock, error) {
	delay := t.retry.InitialDelay
	var lastErr error
	for attempt := 0; ; attempt++ {
		currentSlot, err := t.blockInteractor.GetSlot(ctx)
		if err != nil {
			return nil, err
		}

		ready := slot <= currentSlot
		if ready {
			b, err := t.blockInteractor.GetBlock(ctx, slot)
			if err == nil {
				return b, nil
			}
			if !errors.Is(err, block.ErrBlockNotAvailable) {
				return nil, err
			}
			lastErr = err
		} else {
			lastErr = fmt.Errorf("slot %d is ahead of node's current slot %d", slot, currentSlot)
		}

		if attempt == t.retry.MaxRetries {
			return nil, lastErr
		}

		if ready {
			logging.WarnWithCtx(
				ctx,
				"slot processor trigger: block not yet available, retrying",
				"slot", slot, "attempt", attempt+1, "delay", delay,
			)
		} else {
			logging.WarnWithCtx(
				ctx,
				"slot processor trigger: slot not yet reached by node, waiting",
				"slot", slot, "currentSlot", currentSlot, "attempt", attempt+1, "delay", delay,
			)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
		delay = min(time.Duration(float64(delay)*t.retry.Multiplier), t.retry.MaxDelay)
	}
}

// Supports reports whether e is a SlotEvent.
func (t *SlotProcessorPermanentTrigger) Supports(e event.Event) bool {
	_, ok := e.(event.SlotEvent)
	return ok
}

// toSolanaBlockEvent maps a block.SolanaBlock fetched from a BlockInteractor
// into the SolanaBlockEvent dispatched for it.
func toSolanaBlockEvent(nodeID uuid.UUID, b *block.SolanaBlock) (event.SolanaBlockEvent, error) {
	transactions := make([]event.SolanaTransaction, 0, len(b.Transactions))
	for _, tx := range b.Transactions {
		solanaTx, err := event.NewSolanaTransaction(nodeID, tx.Signature, b.Slot, toInstructionData(tx.Instructions), tx.Accounts, tx.Logs, tx.Err)
		if err != nil {
			return event.SolanaBlockEvent{}, err
		}
		transactions = append(transactions, solanaTx)
	}

	var blockTime *int64
	if b.BlockTime != nil {
		unix := b.BlockTime.Unix()
		blockTime = &unix
	}

	return event.NewSolanaBlockEvent(nodeID, b.Slot, b.Blockhash, blockTime, transactions)
}

// toInstructionData maps a SolanaTransaction's raw Instructions into the
// InstructionData carried by the SolanaTransactionEvent built for it.
func toInstructionData(instructions []block.Instruction) []event.InstructionData {
	data := make([]event.InstructionData, len(instructions))
	for i, instr := range instructions {
		data[i] = event.InstructionData{
			ProgramID: instr.ProgramID,
			Data:      instr.Data,
			Accounts:  instr.Accounts,
			Inner:     instr.Inner,
		}
	}
	return data
}

var _ trigger.PermanentTrigger = (*SlotProcessorPermanentTrigger)(nil)
