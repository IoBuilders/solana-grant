package trigger

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/configurationmanager"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/store"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/store/feature"
)

type LatestBlockStorePermanentTrigger[B event.BlockEvent] struct {
	latestBlockStore   store.LatestBlockStore[B]
	storeConfigmanager configurationmanager.StoreConfigurationManager
}

func NewLatestBlockStorePermanentTrigger[B event.BlockEvent](
	latestBlockStore store.LatestBlockStore[B],
	storeConfigmanager configurationmanager.StoreConfigurationManager,
) *LatestBlockStorePermanentTrigger[B] {
	return &LatestBlockStorePermanentTrigger[B]{
		latestBlockStore:   latestBlockStore,
		storeConfigmanager: storeConfigmanager,
	}
}

func (t *LatestBlockStorePermanentTrigger[B]) Process(ctx context.Context, e event.Event) error {
	latestBlockFeatureConfig, err := getStoreFeatureConfigFromConfigManager[*feature.LatestBlockConfiguration](ctx, t.storeConfigmanager, feature.TypeLatestBlock, e.NodeID())
	if err != nil || latestBlockFeatureConfig == nil {
		return err
	}
	switch ev := e.(type) {
	case B:
		return t.latestBlockStore.Save(ctx, ev)
	default:
		return nil
	}
}

func (t *LatestBlockStorePermanentTrigger[B]) Supports(e event.Event) bool {
	_, ok := e.(B)
	return ok
}

var _ PermanentTrigger = (*LatestBlockStorePermanentTrigger[event.SolanaBlockEvent])(nil)
