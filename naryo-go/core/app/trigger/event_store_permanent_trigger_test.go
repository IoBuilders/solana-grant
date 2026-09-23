//go:build test

package trigger

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/store"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/store/feature"
)

// --- stubs ---

// fakeBlockStore is a hand-written test double for
// store.BlockEventStore[event.SolanaBlockEvent].
type fakeBlockStore struct {
	saved []event.SolanaBlockEvent
	err   error
}

func (s *fakeBlockStore) Save(_ context.Context, data event.SolanaBlockEvent) error {
	if s.err != nil {
		return s.err
	}
	s.saved = append(s.saved, data)
	return nil
}

func (s *fakeBlockStore) Supports(store.Type, reflect.Type) bool { return false }

func (s *fakeBlockStore) GetLatest(context.Context, uuid.UUID) (uint64, error) { return 0, nil }

// fakeTransactionStore is a hand-written test double for
// store.TransactionEventStore[event.SolanaTransactionEvent].
type fakeTransactionStore struct {
	saved []event.SolanaTransactionEvent
	err   error
}

func (s *fakeTransactionStore) Save(_ context.Context, data event.SolanaTransactionEvent) error {
	if s.err != nil {
		return s.err
	}
	s.saved = append(s.saved, data)
	return nil
}

func (s *fakeTransactionStore) Supports(store.Type, reflect.Type) bool { return false }

// fakeContractStore is a hand-written test double for
// store.ContractEventStore[event.SolanaContractEvent].
type fakeContractStore struct {
	saved []event.SolanaContractEvent
	err   error
}

func (s *fakeContractStore) Save(_ context.Context, data event.SolanaContractEvent) error {
	if s.err != nil {
		return s.err
	}
	s.saved = append(s.saved, data)
	return nil
}

func (s *fakeContractStore) Supports(store.Type, reflect.Type) bool { return false }

// stubStoreConfigurationManager is a hand-written test double for
// configurationmanager.StoreConfigurationManager.
type stubStoreConfigurationManager struct {
	configs []store.Configuration
	err     error
}

func (s *stubStoreConfigurationManager) Load(context.Context) ([]store.Configuration, error) {
	return s.configs, s.err
}

func mustSolanaBlockEvent(t *testing.T) event.SolanaBlockEvent {
	t.Helper()
	e, err := event.NewSolanaBlockEvent(uuid.New(), 1, "blockhash", nil, nil)
	require.NoError(t, err)
	return e
}

func mustSolanaTransactionEvent(t *testing.T) event.SolanaTransactionEvent {
	t.Helper()
	e, err := event.NewSolanaTransactionEvent(uuid.New(), "signature", 1, nil, nil, nil, nil)
	require.NoError(t, err)
	return e
}

func mustSolanaContractEvent(t *testing.T) event.SolanaContractEvent {
	t.Helper()
	e, err := event.NewSolanaContractEvent(
		uuid.New(), "programID", "signature", 1,
		[]parameter.ContractEventParameter{}, "eventName", event.ContractEventStatusConfirmed,
	)
	require.NoError(t, err)
	return e
}

// newActiveEventConfig builds an ActiveConfiguration for nodeID whose only
// configured feature is EVENT, routing exactly the given target kinds. It
// goes through the same constructors core/infrastructure/config/store.go uses
// to build one from properties, so the fixture matches real configuration.
func newActiveEventConfig(t *testing.T, nodeID uuid.UUID, targetTypes ...feature.TargetType) store.Configuration {
	t.Helper()
	targets := make([]feature.EventTarget, 0, len(targetTypes))
	for _, targetType := range targetTypes {
		target, err := feature.NewEventTarget(targetType, feature.Destination(""))
		require.NoError(t, err)
		targets = append(targets, target)
	}
	eventConfig, err := feature.NewEventConfiguration(feature.StrategyBlockBased, targets)
	require.NoError(t, err)
	cfg, err := store.NewActiveConfiguration(nodeID, store.TypeGorm, map[feature.Type]feature.Configuration{
		feature.TypeEvent: eventConfig,
	})
	require.NoError(t, err)
	return cfg
}

// newActiveConfigWithoutEventFeature builds an ActiveConfiguration for nodeID
// that configures a feature other than EVENT, so the node is active but has
// nothing to route the event through.
func newActiveConfigWithoutEventFeature(t *testing.T, nodeID uuid.UUID) store.Configuration {
	t.Helper()
	latestBlockConfig, err := feature.NewLatestBlockConfiguration(feature.Destination(""))
	require.NoError(t, err)
	cfg, err := store.NewActiveConfiguration(nodeID, store.TypeGorm, map[feature.Type]feature.Configuration{
		feature.TypeLatestBlock: latestBlockConfig,
	})
	require.NoError(t, err)
	return cfg
}

func mustInactiveConfiguration(t *testing.T, nodeID uuid.UUID) store.Configuration {
	t.Helper()
	cfg, err := store.NewInactiveConfiguration(nodeID)
	require.NoError(t, err)
	return cfg
}

// --- tests ---

func TestEventStorePermanentTrigger_Process_SavesBlockEventToBlockStore(t *testing.T) {
	blockStore := &fakeBlockStore{}
	e := mustSolanaBlockEvent(t)
	configManager := &stubStoreConfigurationManager{
		configs: []store.Configuration{newActiveEventConfig(t, e.NodeID(), feature.TargetTypeBlock)},
	}
	trg := NewEventStorePermanentTrigger[event.SolanaBlockEvent, event.SolanaTransactionEvent, event.SolanaContractEvent](blockStore, &fakeTransactionStore{}, &fakeContractStore{}, configManager)

	err := trg.Process(context.Background(), e)

	require.NoError(t, err)
	assert.Equal(t, []event.SolanaBlockEvent{e}, blockStore.saved)
}

func TestEventStorePermanentTrigger_Process_SavesTransactionEventToTransactionStore(t *testing.T) {
	transactionStore := &fakeTransactionStore{}
	e := mustSolanaTransactionEvent(t)
	configManager := &stubStoreConfigurationManager{
		configs: []store.Configuration{newActiveEventConfig(t, e.NodeID(), feature.TargetTypeTransaction)},
	}
	trg := NewEventStorePermanentTrigger[event.SolanaBlockEvent, event.SolanaTransactionEvent, event.SolanaContractEvent](&fakeBlockStore{}, transactionStore, &fakeContractStore{}, configManager)

	err := trg.Process(context.Background(), e)

	require.NoError(t, err)
	assert.Equal(t, []event.SolanaTransactionEvent{e}, transactionStore.saved)
}

func TestEventStorePermanentTrigger_Process_SavesContractEventToContractStore(t *testing.T) {
	contractStore := &fakeContractStore{}
	e := mustSolanaContractEvent(t)
	configManager := &stubStoreConfigurationManager{
		configs: []store.Configuration{newActiveEventConfig(t, e.NodeID(), feature.TargetTypeContractEvent)},
	}
	trg := NewEventStorePermanentTrigger[event.SolanaBlockEvent, event.SolanaTransactionEvent, event.SolanaContractEvent](&fakeBlockStore{}, &fakeTransactionStore{}, contractStore, configManager)

	err := trg.Process(context.Background(), e)

	require.NoError(t, err)
	assert.Equal(t, []event.SolanaContractEvent{e}, contractStore.saved)
}

func TestEventStorePermanentTrigger_Process_TargetNotConfigured_DoesNothing(t *testing.T) {
	blockStore := &fakeBlockStore{}
	e := mustSolanaBlockEvent(t)
	// The node's EVENT feature is active but only routes transactions, not
	// blocks, so a block event must be silently dropped.
	configManager := &stubStoreConfigurationManager{
		configs: []store.Configuration{newActiveEventConfig(t, e.NodeID(), feature.TargetTypeTransaction)},
	}
	trg := NewEventStorePermanentTrigger[event.SolanaBlockEvent, event.SolanaTransactionEvent, event.SolanaContractEvent](blockStore, &fakeTransactionStore{}, &fakeContractStore{}, configManager)

	err := trg.Process(context.Background(), e)

	require.NoError(t, err)
	assert.Empty(t, blockStore.saved)
}

func TestEventStorePermanentTrigger_Process_NoActiveConfigForNode_DoesNothing(t *testing.T) {
	blockStore := &fakeBlockStore{}
	e := mustSolanaBlockEvent(t)
	configManager := &stubStoreConfigurationManager{
		configs: []store.Configuration{
			newActiveEventConfig(t, uuid.New(), feature.TargetTypeBlock), // a different node
			mustInactiveConfiguration(t, e.NodeID()),                     // this node, but inactive
		},
	}
	trg := NewEventStorePermanentTrigger[event.SolanaBlockEvent, event.SolanaTransactionEvent, event.SolanaContractEvent](blockStore, &fakeTransactionStore{}, &fakeContractStore{}, configManager)

	err := trg.Process(context.Background(), e)

	require.NoError(t, err)
	assert.Empty(t, blockStore.saved)
}

func TestEventStorePermanentTrigger_Process_NoEventFeatureConfigured_DoesNothing(t *testing.T) {
	blockStore := &fakeBlockStore{}
	e := mustSolanaBlockEvent(t)
	configManager := &stubStoreConfigurationManager{
		configs: []store.Configuration{newActiveConfigWithoutEventFeature(t, e.NodeID())},
	}
	trg := NewEventStorePermanentTrigger[event.SolanaBlockEvent, event.SolanaTransactionEvent, event.SolanaContractEvent](blockStore, &fakeTransactionStore{}, &fakeContractStore{}, configManager)

	err := trg.Process(context.Background(), e)

	require.NoError(t, err)
	assert.Empty(t, blockStore.saved)
}

func TestEventStorePermanentTrigger_Process_UnrecognizedEventKind_DoesNothing(t *testing.T) {
	blockStore := &fakeBlockStore{}
	transactionStore := &fakeTransactionStore{}
	contractStore := &fakeContractStore{}
	slotEvent, err := event.NewSlotEvent(uuid.New(), 1, time.Now())
	require.NoError(t, err)
	// Even a node fully configured to route every kind must not save a kind
	// the trigger isn't generic over.
	configManager := &stubStoreConfigurationManager{
		configs: []store.Configuration{newActiveEventConfig(t, slotEvent.NodeID(),
			feature.TargetTypeBlock, feature.TargetTypeTransaction, feature.TargetTypeContractEvent)},
	}
	trg := NewEventStorePermanentTrigger[event.SolanaBlockEvent, event.SolanaTransactionEvent, event.SolanaContractEvent](blockStore, transactionStore, contractStore, configManager)

	err = trg.Process(context.Background(), slotEvent)

	require.NoError(t, err)
	assert.Empty(t, blockStore.saved)
	assert.Empty(t, transactionStore.saved)
	assert.Empty(t, contractStore.saved)
}

func TestEventStorePermanentTrigger_Process_SaveError_IsPropagated(t *testing.T) {
	saveErr := errors.New("save failed")
	blockStore := &fakeBlockStore{err: saveErr}
	e := mustSolanaBlockEvent(t)
	configManager := &stubStoreConfigurationManager{
		configs: []store.Configuration{newActiveEventConfig(t, e.NodeID(), feature.TargetTypeBlock)},
	}
	trg := NewEventStorePermanentTrigger[event.SolanaBlockEvent, event.SolanaTransactionEvent, event.SolanaContractEvent](blockStore, &fakeTransactionStore{}, &fakeContractStore{}, configManager)

	err := trg.Process(context.Background(), e)

	assert.ErrorIs(t, err, saveErr)
}

func TestEventStorePermanentTrigger_Process_ConfigLoadError_IsPropagated(t *testing.T) {
	loadErr := errors.New("load failed")
	blockStore := &fakeBlockStore{}
	configManager := &stubStoreConfigurationManager{err: loadErr}
	trg := NewEventStorePermanentTrigger[event.SolanaBlockEvent, event.SolanaTransactionEvent, event.SolanaContractEvent](blockStore, &fakeTransactionStore{}, &fakeContractStore{}, configManager)

	err := trg.Process(context.Background(), mustSolanaBlockEvent(t))

	assert.ErrorIs(t, err, loadErr)
	assert.Empty(t, blockStore.saved)
}

func TestEventStorePermanentTrigger_Supports(t *testing.T) {
	trg := NewEventStorePermanentTrigger[event.SolanaBlockEvent, event.SolanaTransactionEvent, event.SolanaContractEvent](&fakeBlockStore{}, &fakeTransactionStore{}, &fakeContractStore{}, &stubStoreConfigurationManager{})

	assert.True(t, trg.Supports(mustSolanaBlockEvent(t)))
	assert.True(t, trg.Supports(mustSolanaTransactionEvent(t)))
	assert.True(t, trg.Supports(mustSolanaContractEvent(t)))

	slotEvent, err := event.NewSlotEvent(uuid.New(), 1, time.Now())
	require.NoError(t, err)
	assert.False(t, trg.Supports(slotEvent))
}
