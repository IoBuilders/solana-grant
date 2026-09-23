//go:build test

package slotprocessor

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/interactor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/trigger"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/block"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/common"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

// --- stubs ---

type stubEvent struct{}

func (stubEvent) EventType() event.Type { return event.TypeBlock }
func (stubEvent) NodeID() uuid.UUID     { return uuid.Nil }

// fakeBlockInteractor is a hand-written test double for interactor.BlockInteractor.
// Each GetBlock call pops one error off errs, if any remain, before falling
// back to err (returned on every subsequent call) or block; each GetSlot call
// pops one value off slots, if any remain, before falling back to slot. This
// lets tests simulate a block that becomes available only after a number of
// failed attempts, and a node's current slot that advances across calls.
type fakeBlockInteractor struct {
	block *block.SolanaBlock
	err   error
	errs  []error
	calls int

	slot      uint64
	slots     []uint64
	slotErr   error
	slotCalls int
}

func (i *fakeBlockInteractor) Type() interactor.Type { return interactor.TypeBlock }

func (i *fakeBlockInteractor) GetBlock(context.Context, uint64) (*block.SolanaBlock, error) {
	i.calls++
	if len(i.errs) > 0 {
		err := i.errs[0]
		i.errs = i.errs[1:]
		return nil, err
	}
	if i.err != nil {
		return nil, i.err
	}
	return i.block, nil
}

func (i *fakeBlockInteractor) GetSlot(context.Context) (uint64, error) {
	i.slotCalls++
	if i.slotErr != nil {
		return 0, i.slotErr
	}
	if len(i.slots) > 0 {
		s := i.slots[0]
		i.slots = i.slots[1:]
		return s, nil
	}
	return i.slot, nil
}

// fakeDispatcher is a hand-written test double for dispatch.Dispatcher.
type fakeDispatcher struct {
	dispatched []event.Event
}

func (d *fakeDispatcher) Dispatch(_ context.Context, e event.Event) {
	d.dispatched = append(d.dispatched, e)
}

func (d *fakeDispatcher) AddTrigger(trigger.Trigger) {}

func (d *fakeDispatcher) RemoveExpired() {}

func mustSlotEvent(t *testing.T, nodeID uuid.UUID, slot uint64) event.SlotEvent {
	t.Helper()
	e, err := event.NewSlotEvent(nodeID, slot, time.Now())
	require.NoError(t, err)
	return e
}

// fastRetry builds a RetryConfiguration with millisecond-scale delays, so
// tests exercising the retry path don't wait on real backoff durations.
func fastRetry(t *testing.T, maxRetries int) *common.RetryConfiguration {
	t.Helper()
	r, err := common.NewRetryConfiguration(maxRetries, time.Millisecond, 5*time.Millisecond, 2)
	require.NoError(t, err)
	return r
}

// --- tests ---

func TestSlotProcessorPermanentTrigger_Process_DispatchesBlockEvent(t *testing.T) {
	nodeID := uuid.New()
	blockTime := time.Unix(1_700_000_000, 0)
	b := &block.SolanaBlock{
		Slot:      42,
		Blockhash: "hash",
		BlockTime: &blockTime,
		Transactions: []block.SolanaTransaction{
			{
				Signature:    "sig",
				Instructions: []block.Instruction{{ProgramID: "prog", Data: []byte{1, 2}, Accounts: []string{"acc"}}},
				Logs:         []string{"log line"},
			},
		},
	}
	blockInteractor := &fakeBlockInteractor{block: b, slot: 42}
	dispatcher := &fakeDispatcher{}
	trg := NewSlotProcessorPermanentTrigger(blockInteractor, dispatcher, common.DefaultRetryConfiguration())

	err := trg.Process(context.Background(), mustSlotEvent(t, nodeID, 42))

	require.NoError(t, err)
	require.Len(t, dispatcher.dispatched, 1)

	blockEvent, ok := dispatcher.dispatched[0].(event.SolanaBlockEvent)
	require.True(t, ok)
	assert.Equal(t, nodeID, blockEvent.NodeID())
	assert.Equal(t, uint64(42), blockEvent.Slot)
	assert.Equal(t, "hash", blockEvent.Blockhash)
	require.NotNil(t, blockEvent.BlockTime)
	assert.Equal(t, blockTime.Unix(), *blockEvent.BlockTime)

	transactions := blockEvent.Transactions()
	require.Len(t, transactions, 1)
	tx, ok := transactions[0].(event.SolanaTransaction)
	require.True(t, ok)
	assert.Equal(t, "sig", tx.Signature)
	assert.Equal(t, uint64(42), tx.Slot)
	assert.Equal(t, []event.InstructionData{{ProgramID: "prog", Data: []byte{1, 2}, Accounts: []string{"acc"}}}, tx.Instructions)
	assert.Equal(t, []string{"log line"}, tx.Logs)
}

func TestSlotProcessorPermanentTrigger_Process_GetBlockError_ReturnsError(t *testing.T) {
	blockInteractor := &fakeBlockInteractor{err: errors.New("rpc failed"), slot: 42}
	dispatcher := &fakeDispatcher{}
	trg := NewSlotProcessorPermanentTrigger(blockInteractor, dispatcher, common.DefaultRetryConfiguration())

	err := trg.Process(context.Background(), mustSlotEvent(t, uuid.New(), 42))

	require.Error(t, err)
	assert.Empty(t, dispatcher.dispatched)
	assert.Equal(t, 1, blockInteractor.calls, "a generic error must not be retried")
}

func TestSlotProcessorPermanentTrigger_Process_SlotSkipped_ReturnsNilWithoutDispatch(t *testing.T) {
	blockInteractor := &fakeBlockInteractor{err: fmt.Errorf("rpc: %w", block.ErrSlotSkipped), slot: 42}
	dispatcher := &fakeDispatcher{}
	trg := NewSlotProcessorPermanentTrigger(blockInteractor, dispatcher, common.DefaultRetryConfiguration())

	err := trg.Process(context.Background(), mustSlotEvent(t, uuid.New(), 42))

	require.NoError(t, err)
	assert.Empty(t, dispatcher.dispatched)
	assert.Equal(t, 1, blockInteractor.calls, "a skipped slot must not be retried")
}

func TestSlotProcessorPermanentTrigger_Process_BlockNotAvailable_RetriesThenSucceeds(t *testing.T) {
	nodeID := uuid.New()
	b := &block.SolanaBlock{Slot: 42, Blockhash: "hash"}
	blockInteractor := &fakeBlockInteractor{
		errs: []error{
			fmt.Errorf("rpc: %w", block.ErrBlockNotAvailable),
			fmt.Errorf("rpc: %w", block.ErrBlockNotAvailable),
		},
		block: b,
		slot:  42,
	}
	dispatcher := &fakeDispatcher{}
	trg := NewSlotProcessorPermanentTrigger(blockInteractor, dispatcher, fastRetry(t, 3))

	err := trg.Process(context.Background(), mustSlotEvent(t, nodeID, 42))

	require.NoError(t, err)
	require.Len(t, dispatcher.dispatched, 1)
	assert.Equal(t, 3, blockInteractor.calls)
}

// Requested slot 42 never exceeds the node's current slot (always 42), so
// GetBlock is called on every attempt; when it persistently reports
// BlockNotAvailable, exhausting the retry budget must surface as an error,
// not a silent skip — only an actual SlotSkipped response does that.
func TestSlotProcessorPermanentTrigger_Process_BlockNotAvailable_ExhaustsRetries_ReturnsError(t *testing.T) {
	blockInteractor := &fakeBlockInteractor{err: fmt.Errorf("rpc: %w", block.ErrBlockNotAvailable), slot: 42}
	dispatcher := &fakeDispatcher{}
	trg := NewSlotProcessorPermanentTrigger(blockInteractor, dispatcher, fastRetry(t, 2))

	err := trg.Process(context.Background(), mustSlotEvent(t, uuid.New(), 42))

	require.Error(t, err)
	assert.ErrorIs(t, err, block.ErrBlockNotAvailable)
	assert.Empty(t, dispatcher.dispatched)
	assert.Equal(t, 3, blockInteractor.calls, "1 initial attempt + 2 retries")
	assert.Equal(t, 3, blockInteractor.slotCalls, "GetSlot is checked before every GetBlock attempt")
}

// While the requested slot is still ahead of the node's current slot, the
// trigger must wait without even calling GetBlock. Once the node catches up
// (simulated by the slots sequence), it proceeds to fetch the block.
func TestSlotProcessorPermanentTrigger_Process_FutureSlot_WaitsThenSucceedsWithoutPrematureGetBlock(t *testing.T) {
	nodeID := uuid.New()
	b := &block.SolanaBlock{Slot: 42, Blockhash: "hash"}
	blockInteractor := &fakeBlockInteractor{
		slots: []uint64{10, 20, 50}, // catches up to (and passes) slot 42 on the 3rd check
		block: b,
	}
	dispatcher := &fakeDispatcher{}
	trg := NewSlotProcessorPermanentTrigger(blockInteractor, dispatcher, fastRetry(t, 5))

	err := trg.Process(context.Background(), mustSlotEvent(t, nodeID, 42))

	require.NoError(t, err)
	require.Len(t, dispatcher.dispatched, 1)
	assert.Equal(t, 1, blockInteractor.calls, "GetBlock must only be called once the slot is reached")
	assert.Equal(t, 3, blockInteractor.slotCalls)
}

// The requested slot never stops being ahead of the node's current slot, so
// GetBlock must never be called; once the retry budget is exhausted, Process
// returns an error rather than silently giving up.
func TestSlotProcessorPermanentTrigger_Process_FutureSlot_ExhaustsRetries_ReturnsErrorWithoutCallingGetBlock(t *testing.T) {
	blockInteractor := &fakeBlockInteractor{slot: 0}
	dispatcher := &fakeDispatcher{}
	trg := NewSlotProcessorPermanentTrigger(blockInteractor, dispatcher, fastRetry(t, 2))

	err := trg.Process(context.Background(), mustSlotEvent(t, uuid.New(), 42))

	require.Error(t, err)
	assert.Empty(t, dispatcher.dispatched)
	assert.Equal(t, 0, blockInteractor.calls, "GetBlock must never be called while the slot is still in the future")
	assert.Equal(t, 3, blockInteractor.slotCalls, "1 initial check + 2 retries")
}

// A failed GetSlot check is a real, unrelated failure (not evidence of a
// pending or skipped slot), so it must surface immediately without
// consuming the retry budget or calling GetBlock.
func TestSlotProcessorPermanentTrigger_Process_GetSlotFails_ReturnsErrorImmediately(t *testing.T) {
	blockInteractor := &fakeBlockInteractor{slotErr: errors.New("get slot rpc failed")}
	dispatcher := &fakeDispatcher{}
	trg := NewSlotProcessorPermanentTrigger(blockInteractor, dispatcher, fastRetry(t, 3))

	err := trg.Process(context.Background(), mustSlotEvent(t, uuid.New(), 42))

	require.Error(t, err)
	assert.Empty(t, dispatcher.dispatched)
	assert.Equal(t, 0, blockInteractor.calls)
	assert.Equal(t, 1, blockInteractor.slotCalls, "a GetSlot failure must not be retried")
}

func TestSlotProcessorPermanentTrigger_Process_MappingError_ReturnsError(t *testing.T) {
	b := &block.SolanaBlock{Slot: 42, Blockhash: ""}
	blockInteractor := &fakeBlockInteractor{block: b, slot: 42}
	dispatcher := &fakeDispatcher{}
	trg := NewSlotProcessorPermanentTrigger(blockInteractor, dispatcher, common.DefaultRetryConfiguration())

	err := trg.Process(context.Background(), mustSlotEvent(t, uuid.New(), 42))

	require.Error(t, err)
	assert.Empty(t, dispatcher.dispatched)
}

func TestSlotProcessorPermanentTrigger_Process_IgnoresUnsupportedEvent(t *testing.T) {
	blockInteractor := &fakeBlockInteractor{}
	dispatcher := &fakeDispatcher{}
	trg := NewSlotProcessorPermanentTrigger(blockInteractor, dispatcher, common.DefaultRetryConfiguration())

	err := trg.Process(context.Background(), stubEvent{})

	require.NoError(t, err)
	assert.Empty(t, dispatcher.dispatched)
}

func TestSlotProcessorPermanentTrigger_Supports(t *testing.T) {
	trg := NewSlotProcessorPermanentTrigger(&fakeBlockInteractor{}, &fakeDispatcher{}, common.DefaultRetryConfiguration())

	assert.True(t, trg.Supports(mustSlotEvent(t, uuid.New(), 42)))
	assert.False(t, trg.Supports(stubEvent{}))
}
