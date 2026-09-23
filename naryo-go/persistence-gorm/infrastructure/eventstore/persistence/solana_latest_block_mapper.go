package eventstorepersistence

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

func FromSolanaLatestBlockDomain(domain event.SolanaBlockEvent) (SolanaLatestBlock, error) {
	return SolanaLatestBlock{
		NodeID:    domain.NodeID(),
		Slot:      domain.Slot,
		Blockhash: domain.Blockhash,
		BlockTime: domain.BlockTime,
	}, nil
}
