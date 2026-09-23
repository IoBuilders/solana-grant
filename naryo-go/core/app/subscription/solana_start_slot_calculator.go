package subscription

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/interactor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/store"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/node"
)

type SolanaStartSlotCalculator struct {
	n                *node.Node
	blockInteractor  interactor.BlockInteractor
	latestBlockStore store.LatestBlockStore[event.SolanaBlockEvent]
}

func NewSolanaStartSlotCalculator(n *node.Node, blockInteractor interactor.BlockInteractor, latestBlockStore store.LatestBlockStore[event.SolanaBlockEvent]) *SolanaStartSlotCalculator {
	return &SolanaStartSlotCalculator{
		n:                n,
		blockInteractor:  blockInteractor,
		latestBlockStore: latestBlockStore,
	}
}

func (c *SolanaStartSlotCalculator) GetStartSlot(ctx context.Context) (uint64, error) {
	// TODO include initial block, sync block limit and rest of configuration to retrieve start slot
	startSlot, err := c.latestBlockStore.Get(ctx, c.n.ID)
	if err != nil {
		return 0, err
	}
	if startSlot == 0 {
		startSlot, err = c.blockInteractor.GetSlot(ctx) // Latest slot in solana network
	} else {
		startSlot++
	}
	return startSlot, err
}
