package transactionprocessor

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/configurationmanager"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter"
)

// TransactionHandler processes a single BlockTransaction extracted from a
// BlockEvent against the currently active TransactionFilters. Concrete
// TransactionProcessorPermanentTrigger implementations supply their own
// chain-specific matching and dispatch rule.
type TransactionHandler func(ctx context.Context, tx event.BlockTransaction, filters []*filter.TransactionFilter)

// TransactionProcessorPermanentTrigger is the base processing contract
// shared by every chain-specific transaction processor: a block with no
// transactions is skipped, the current active TRANSACTION filters are
// loaded once per block, and every transaction is handed — together with
// that filter set — to a TransactionHandler.
//
// It is not meant to be used on its own: chain-specific concrete types (e.g.
// SolanaTransactionProcessorPermanentTrigger) embed it and call process with
// their own TransactionHandler.
type TransactionProcessorPermanentTrigger struct {
	filterManager configurationmanager.FilterConfigurationManager
}

// process loads the currently active TRANSACTION filters and hands every
// one of e's transactions, together with that filter set, to handle. A
// block with no transactions is skipped entirely, without loading filters.
func (t TransactionProcessorPermanentTrigger) process(ctx context.Context, e event.BlockEvent, handle TransactionHandler) error {
	transactions := e.Transactions()
	if len(transactions) == 0 {
		return nil
	}

	filters, err := t.activeTransactionFilters(ctx, e.NodeID())
	if err != nil {
		return fmt.Errorf("transaction processor trigger: failed to load active filters: %w", err)
	}
	if len(filters) == 0 {
		return nil
	}

	for _, tx := range transactions {
		handle(ctx, tx, filters)
	}
	return nil
}

// activeTransactionFilters returns every currently configured Filter of
// type TRANSACTION belonging to nodeID, as its concrete *filter.TransactionFilter.
func (t TransactionProcessorPermanentTrigger) activeTransactionFilters(ctx context.Context, nodeID uuid.UUID) ([]*filter.TransactionFilter, error) {
	filters, err := t.filterManager.Load(ctx)
	if err != nil {
		return nil, err
	}

	transactionFilters := make([]*filter.TransactionFilter, 0, len(filters))
	for _, f := range filters {
		if f.NodeID() != nodeID {
			continue
		}
		tf, ok := f.(*filter.TransactionFilter)
		if !ok {
			continue
		}
		transactionFilters = append(transactionFilters, tf)
	}
	return transactionFilters, nil
}
