package blockprocessor

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/configurationmanager"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter"
)

// TransactionHandler processes a single BlockTransaction extracted from a
// BlockEvent against the currently active EventFilters. Concrete
// BlockProcessorPermanentTrigger implementations supply their own
// chain-specific rule for matching a transaction's contents against filters,
// decoding whatever matches, and dispatching the resulting ContractEvent.
// Per-transaction failures (a single bad log, a single decode error) are the
// handler's own responsibility to isolate and log — one failure must not
// stop the rest of the block from being processed.
type TransactionHandler func(ctx context.Context, nodeID uuid.UUID, tx event.BlockTransaction, filters []*filter.EventFilter)

// BlockProcessorPermanentTrigger is the base processing contract shared by
// every chain-specific block processor: a block with no transactions is
// skipped, the current active EVENT filters are loaded once per block, and
// every transaction is handed — together with that filter set — to a
// TransactionHandler.
//
// It is not meant to be used on its own: chain-specific concrete types (e.g.
// SolanaBlockProcessorPermanentTrigger) embed it and call process with their
// own TransactionHandler.
type BlockProcessorPermanentTrigger struct {
	filterManager configurationmanager.FilterConfigurationManager
}

// process loads the currently active EVENT filters and hands every one of
// e's transactions, together with that filter set, to handle. A block with
// no transactions is skipped entirely, without loading filters.
func (t BlockProcessorPermanentTrigger) process(ctx context.Context, e event.BlockEvent, handle TransactionHandler) error {
	transactions := e.Transactions()
	if len(transactions) == 0 {
		return nil
	}

	nodeID := e.NodeID()
	filters, err := t.activeEventFilters(ctx, nodeID)
	if err != nil {
		return fmt.Errorf("block processor trigger: failed to load active filters: %w", err)
	}
	if len(filters) == 0 {
		return nil
	}

	for _, tx := range transactions {
		handle(ctx, nodeID, tx, filters)
	}
	return nil
}

// activeEventFilters returns every currently configured Filter of type EVENT
// belonging to nodeID, as its concrete *filter.EventFilter.
func (t BlockProcessorPermanentTrigger) activeEventFilters(ctx context.Context, nodeID uuid.UUID) ([]*filter.EventFilter, error) {
	filters, err := t.filterManager.Load(ctx)
	if err != nil {
		return nil, err
	}

	eventFilters := make([]*filter.EventFilter, 0, len(filters))
	for _, f := range filters {
		if f.NodeID() != nodeID {
			continue
		}
		ef, ok := f.(*filter.EventFilter)
		if !ok {
			continue
		}
		eventFilters = append(eventFilters, ef)
	}
	return eventFilters, nil
}
