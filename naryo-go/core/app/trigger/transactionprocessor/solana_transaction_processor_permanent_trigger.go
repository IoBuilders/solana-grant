package transactionprocessor

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/configurationmanager"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/dispatch"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/logging"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/trigger"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter"
)

// SolanaTransactionProcessorPermanentTrigger is the Solana
// TransactionProcessorPermanentTrigger: for every transaction in a
// processed BlockEvent, it dispatches the transaction once for every active
// TransactionFilter that matches it. It ignores every other event.Event.
// Before dispatch, a failed transaction is decorated with its
// DecodedError result so that every consumer of the dispatched event --
// not just one broadcaster type -- receives the decoded reason.
type SolanaTransactionProcessorPermanentTrigger struct {
	TransactionProcessorPermanentTrigger

	dispatcher      dispatch.Dispatcher
	registryManager configurationmanager.ProgramErrorRegistryConfigurationManager
}

// NewSolanaTransactionProcessorPermanentTrigger builds a
// SolanaTransactionProcessorPermanentTrigger backed by filterManager,
// dispatcher, and registryManager.
func NewSolanaTransactionProcessorPermanentTrigger(
	filterManager configurationmanager.FilterConfigurationManager,
	dispatcher dispatch.Dispatcher,
	registryManager configurationmanager.ProgramErrorRegistryConfigurationManager,
) *SolanaTransactionProcessorPermanentTrigger {
	return &SolanaTransactionProcessorPermanentTrigger{
		TransactionProcessorPermanentTrigger: TransactionProcessorPermanentTrigger{filterManager: filterManager},
		dispatcher:                           dispatcher,
		registryManager:                      registryManager,
	}
}

// Process dispatches every transaction in e once per active
// TransactionFilter that matches it. Process is only ever called with a
// BlockEvent, since Supports rejects every other event.Event.
func (t *SolanaTransactionProcessorPermanentTrigger) Process(ctx context.Context, e event.Event) error {
	blockEvent, ok := e.(event.BlockEvent)
	if !ok {
		return nil
	}

	return t.process(ctx, blockEvent, t.processTransaction)
}

// Supports reports whether e is a BlockEvent.
func (t *SolanaTransactionProcessorPermanentTrigger) Supports(e event.Event) bool {
	_, ok := e.(event.BlockEvent)
	return ok
}

// processTransaction dispatches the SolanaTransactionEvent built from tx
// once for every filter that matches it.
// A failed transaction is decorated with its decoded reason (see
// decodeError) before dispatch; the decoration never blocks or skips
// dispatch, it just falls back to the undecorated tx.

// processTransaction dispatches the SolanaTransactionEvent built from tx
// once for every filter that matches it. tx that doesn't assert to a
// SolanaTransaction (i.e. isn't Solana) is never dispatched.
func (t *SolanaTransactionProcessorPermanentTrigger) processTransaction(ctx context.Context, tx event.BlockTransaction, filters []*filter.TransactionFilter) {
	solanaTx, ok := tx.(event.SolanaTransaction)
	if !ok {
		return
	}

	dispatchTx := t.decodeError(ctx, solanaTx.ToSolanaTransactionEvent())
	for _, f := range filters {
		if !f.Matches(tx) {
			continue
		}
		t.dispatcher.Dispatch(ctx, dispatchTx)
	}
}

// decodeError returns tx decorated with its DecodedError result when tx is
// a failed SolanaTransactionEvent, or tx unchanged otherwise -- including
// when the registry fails to load or DecodedError finds nothing to report.
// Decoding never blocks or fails dispatch.
func (t *SolanaTransactionProcessorPermanentTrigger) decodeError(ctx context.Context, tx event.TransactionEvent) event.TransactionEvent {
	solanaTx, ok := tx.(event.SolanaTransactionEvent)
	if !ok || !solanaTx.HasError() {
		return tx
	}

	registry, err := t.registryManager.Load(ctx)
	if err != nil {
		logging.ErrorWithCtx(ctx, "failed to load program error registry", "error", err)
		return tx
	}

	reason, ok := solanaTx.DecodedError(registry)
	if !ok {
		return tx
	}
	return solanaTx.WithDecodedReason(reason)
}

var _ trigger.PermanentTrigger = (*SolanaTransactionProcessorPermanentTrigger)(nil)
