package trigger

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/configurationmanager"
	appstore "gitlab.com/iobuilders/projects/eng/naryo-go/core/app/store"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/store/feature"
)

// EventStorePermanentTrigger persists every event it processes through the
// store matching its concrete kind. A Node with nothing configured for a
// given kind is wired with a store.Noop*EventStore instead, so this trigger
// depends on all three ports unconditionally rather than checking whether a
// store is actually configured.
type EventStorePermanentTrigger[B event.BlockEvent, T event.TransactionEvent, C event.ContractEvent] struct {
	blockStore         appstore.BlockEventStore[B]
	transactionStore   appstore.TransactionEventStore[T]
	contractStore      appstore.ContractEventStore[C]
	storeConfigmanager configurationmanager.StoreConfigurationManager
}

// NewEventStorePermanentTrigger builds an EventStorePermanentTrigger that
// saves each event kind through its own store.
func NewEventStorePermanentTrigger[B event.BlockEvent, T event.TransactionEvent, C event.ContractEvent](
	blockStore appstore.BlockEventStore[B],
	transactionStore appstore.TransactionEventStore[T],
	contractStore appstore.ContractEventStore[C],
	storeConfigmanager configurationmanager.StoreConfigurationManager,
) *EventStorePermanentTrigger[B, T, C] {
	return &EventStorePermanentTrigger[B, T, C]{
		blockStore:         blockStore,
		transactionStore:   transactionStore,
		contractStore:      contractStore,
		storeConfigmanager: storeConfigmanager,
	}
}

// Process saves e through the store matching its concrete kind. Any error
// from the store is returned as-is: the Dispatcher that invoked Process
// logs it and isolates it from every other registered Trigger.
func (t *EventStorePermanentTrigger[B, T, C]) Process(ctx context.Context, e event.Event) error {
	eventFeatureConfig, err := getStoreFeatureConfigFromConfigManager[*feature.EventConfiguration](ctx, t.storeConfigmanager, feature.TypeEvent, e.NodeID())
	if err != nil || eventFeatureConfig == nil {
		return err
	}
	switch ev := e.(type) {
	case B:
		if _, ok := eventFeatureConfig.Target(feature.TargetTypeBlock); ok {
			err = t.blockStore.Save(ctx, ev)
		}
	case T:
		if _, ok := eventFeatureConfig.Target(feature.TargetTypeTransaction); ok {
			err = t.transactionStore.Save(ctx, ev)
		}
	case C:
		if _, ok := eventFeatureConfig.Target(feature.TargetTypeContractEvent); ok {
			err = t.contractStore.Save(ctx, ev)
		}
	}
	return err
}

// Supports reports whether e is a kind this trigger persists.
func (t *EventStorePermanentTrigger[B, T, C]) Supports(e event.Event) bool {
	switch e.(type) {
	case B, T, C:
		return true
	default:
		return false
	}
}

var _ PermanentTrigger = (*EventStorePermanentTrigger[event.SolanaBlockEvent, event.SolanaTransactionEvent, event.SolanaContractEvent])(nil)
