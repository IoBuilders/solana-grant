package eventstorepersistence

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

func FromSolanaBlockEventDomain(domain event.SolanaBlockEvent) (SolanaBlockEvent, error) {
	transactions := make([]SolanaBlockTransaction, 0, len(domain.Transactions()))
	for _, transaction := range domain.Transactions() {
		transactionPersistence, err := FromSolanaTransactionDomain(transaction.(event.SolanaTransaction))
		if err != nil {
			return SolanaBlockEvent{}, err
		}
		transactions = append(transactions, *transactionPersistence)
	}
	return SolanaBlockEvent{
		NodeID:       domain.NodeID(),
		Slot:         domain.Slot,
		Blockhash:    domain.Blockhash,
		BlockTime:    domain.BlockTime,
		Transactions: transactions,
	}, nil
}

func ToSolanaBlockEventDomain(persistence SolanaBlockEvent) (event.SolanaBlockEvent, error) {
	transactions := make([]event.SolanaTransaction, 0, len(persistence.Transactions))
	for _, transaction := range persistence.Transactions {
		transactionDomain, err := ToSolanaTransactionDomain(transaction)
		if err != nil {
			return event.SolanaBlockEvent{}, err
		}
		transactions = append(transactions, transactionDomain)
	}
	return event.NewSolanaBlockEvent(persistence.NodeID, persistence.Slot, persistence.Blockhash, persistence.BlockTime, transactions)
}
