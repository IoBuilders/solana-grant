package blockprocessor

import (
	"context"
	"slices"

	"github.com/google/uuid"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/configurationmanager"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/logging"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/trigger"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/decoder"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter/solana"
)

// SolanaBlockProcessorPermanentTrigger is the Solana BlockProcessorPermanentTrigger:
// for every instruction of every transaction in a processed BlockEvent, it
// matches the invoked program against the active EventFilters and, for each
// match, decodes the instruction via Decoder and dispatches a ContractEvent
// for every payload that decodes successfully. It ignores every other
// event.Event.
type SolanaBlockProcessorPermanentTrigger struct {
	BlockProcessorPermanentTrigger

	decoder decoder.Decoder
	helper  *ContractEventDispatcherHelper
}

// NewSolanaBlockProcessorPermanentTrigger builds a
// SolanaBlockProcessorPermanentTrigger backed by filterManager, decoder and
// helper.
func NewSolanaBlockProcessorPermanentTrigger(
	filterManager configurationmanager.FilterConfigurationManager,
	decoder decoder.Decoder,
	helper *ContractEventDispatcherHelper,
) *SolanaBlockProcessorPermanentTrigger {
	return &SolanaBlockProcessorPermanentTrigger{
		BlockProcessorPermanentTrigger: BlockProcessorPermanentTrigger{filterManager: filterManager},
		decoder:                        decoder,
		helper:                         helper,
	}
}

// Process matches, decodes and dispatches ContractEvents for e's
// transactions. Process is only ever called with a BlockEvent, since
// Supports rejects every other event.Event.
func (t *SolanaBlockProcessorPermanentTrigger) Process(ctx context.Context, e event.Event) error {
	blockEvent, ok := e.(event.BlockEvent)
	if !ok {
		return nil
	}

	return t.process(ctx, blockEvent, t.processTransaction)
}

// Supports reports whether e is a BlockEvent.
func (t *SolanaBlockProcessorPermanentTrigger) Supports(e event.Event) bool {
	_, ok := e.(event.BlockEvent)
	return ok
}

// processTransaction matches the program invoked by each of tx's
// instructions against filters and decodes every match.
func (t *SolanaBlockProcessorPermanentTrigger) processTransaction(ctx context.Context, nodeID uuid.UUID, tx event.BlockTransaction, filters []*filter.EventFilter) {
	solanaTx, ok := tx.(event.SolanaTransaction)
	if !ok {
		return
	}

	for _, instruction := range solanaTx.Instructions {
		for _, f := range filters {
			if !matchesProgram(f, instruction.ProgramID) {
				continue
			}
			if !matchesAccountAddress(f, instruction.Accounts) {
				continue
			}
			t.decodeAndDispatch(ctx, nodeID, solanaTx, f, instruction)
		}
	}
}

// decodeAndDispatch decodes instruction against f's Specification and
// dispatches a ContractEvent if it decodes successfully. A nil parameters
// slice from Decoder is a non-match, not an error, per the Decoder contract.
// A decode error or a dispatch failure is logged and does not prevent the
// remaining instructions/filters from being processed.
func (t *SolanaBlockProcessorPermanentTrigger) decodeAndDispatch(ctx context.Context, nodeID uuid.UUID, tx event.SolanaTransaction, f *filter.EventFilter, instruction event.InstructionData) {
	raw := decoder.RawPayload{
		ProgramID: instruction.ProgramID,
		Data:      instruction.Data,
		Accounts:  instruction.Accounts,
	}

	parameters, err := t.decoder.Decode(f.Specification, raw)
	if err != nil {
		logging.ErrorWithCtx(
			ctx,
			"solana block processor trigger: failed to decode payload",
			"filter", f.Name().String(),
			"signature", tx.Signature,
			"error", err,
		)
		return
	}
	if parameters == nil {
		return
	}

	eventName := f.Specification.EventName()
	if err := t.helper.Execute(ctx, nodeID, instruction.ProgramID, tx.Signature, tx.Slot, parameters, eventName, event.ContractEventStatusConfirmed); err != nil {
		logging.ErrorWithCtx(
			ctx,
			"solana block processor trigger: failed to dispatch contract event",
			"filter", f.Name().String(),
			"signature", tx.Signature,
			"error", err,
		)
	}
}

// matchesProgram reports whether f targets programID: every GLOBAL-scope
// filter matches any program, while a CONTRACT-scope filter matches only its
// own ContractAddress.
func matchesProgram(f *filter.EventFilter, programID string) bool {
	if f.Scope == filter.ScopeGlobal {
		return true
	}
	return f.ContractAddress != nil && *f.ContractAddress == programID
}

// matchesAccountAddress reports whether accounts contains the identifying
// account address declared on f's Specification: an AnchorSpecification's
// AccountAddress, or an SplNativeSpecification's MintAddress. A program is
// deployed once but drives many accounts/instances (many token mints all
// running through the same SPL Token program, many PDAs under the same
// Anchor program); this narrows a match down to the specific instance the
// filter targets, the Solana analogue of an Ethereum contract address.
func matchesAccountAddress(f *filter.EventFilter, accounts []string) bool {
	switch spec := f.Specification.(type) {
	case *solana.AnchorSpecification:
		return spec.AccountAddress == nil || slices.Contains(accounts, *spec.AccountAddress)
	case solana.SplNativeSpecification:
		// A nil MintAddress means the filter matches regardless of mint —
		// the only option for instructions whose raw data carries no mint
		// account at all (see SplNativeDecoder).
		return spec.MintAddress == nil || slices.Contains(accounts, *spec.MintAddress)
	default:
		return false
	}
}

var _ trigger.PermanentTrigger = (*SolanaBlockProcessorPermanentTrigger)(nil)
