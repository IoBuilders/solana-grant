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
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/store"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/store/feature"
)

// fakeLatestBlockStore is a hand-written test double for
// store.LatestBlockStore[event.SolanaBlockEvent] (core/app/store).
type fakeLatestBlockStore struct {
	saved []event.SolanaBlockEvent
	err   error
}

func (s *fakeLatestBlockStore) Save(_ context.Context, data event.SolanaBlockEvent) error {
	if s.err != nil {
		return s.err
	}
	s.saved = append(s.saved, data)
	return nil
}

func (s *fakeLatestBlockStore) Supports(store.Type, reflect.Type) bool { return false }

func (s *fakeLatestBlockStore) Get(context.Context, uuid.UUID) (uint64, error) { return 0, nil }

func TestLatestBlockStorePermanentTrigger_Process_SavesBlockEventToLatestBlockStore(t *testing.T) {
	latestBlockStore := &fakeLatestBlockStore{}
	e := mustSolanaBlockEvent(t)
	// newActiveConfigWithoutEventFeature configures LATEST_BLOCK (and nothing
	// else), which is exactly the feature this trigger looks for.
	configManager := &stubStoreConfigurationManager{
		configs: []store.Configuration{newActiveConfigWithoutEventFeature(t, e.NodeID())},
	}
	trg := NewLatestBlockStorePermanentTrigger[event.SolanaBlockEvent](latestBlockStore, configManager)

	err := trg.Process(context.Background(), e)

	require.NoError(t, err)
	assert.Equal(t, []event.SolanaBlockEvent{e}, latestBlockStore.saved)
}

func TestLatestBlockStorePermanentTrigger_Process_NoActiveConfigForNode_DoesNothing(t *testing.T) {
	latestBlockStore := &fakeLatestBlockStore{}
	e := mustSolanaBlockEvent(t)
	configManager := &stubStoreConfigurationManager{
		configs: []store.Configuration{
			newActiveConfigWithoutEventFeature(t, uuid.New()), // a different node
			mustInactiveConfiguration(t, e.NodeID()),          // this node, but inactive
		},
	}
	trg := NewLatestBlockStorePermanentTrigger[event.SolanaBlockEvent](latestBlockStore, configManager)

	err := trg.Process(context.Background(), e)

	require.NoError(t, err)
	assert.Empty(t, latestBlockStore.saved)
}

func TestLatestBlockStorePermanentTrigger_Process_NoLatestBlockFeatureConfigured_DoesNothing(t *testing.T) {
	latestBlockStore := &fakeLatestBlockStore{}
	e := mustSolanaBlockEvent(t)
	// The node is active, but only has an EVENT feature, not LATEST_BLOCK.
	configManager := &stubStoreConfigurationManager{
		configs: []store.Configuration{newActiveEventConfig(t, e.NodeID(), feature.TargetTypeBlock)},
	}
	trg := NewLatestBlockStorePermanentTrigger[event.SolanaBlockEvent](latestBlockStore, configManager)

	err := trg.Process(context.Background(), e)

	require.NoError(t, err)
	assert.Empty(t, latestBlockStore.saved)
}

func TestLatestBlockStorePermanentTrigger_Process_UnrecognizedEventKind_DoesNothing(t *testing.T) {
	latestBlockStore := &fakeLatestBlockStore{}
	slotEvent, err := event.NewSlotEvent(uuid.New(), 1, time.Now())
	require.NoError(t, err)
	// Even with LATEST_BLOCK configured for this node, a SlotEvent is not a B.
	configManager := &stubStoreConfigurationManager{
		configs: []store.Configuration{newActiveConfigWithoutEventFeature(t, slotEvent.NodeID())},
	}
	trg := NewLatestBlockStorePermanentTrigger[event.SolanaBlockEvent](latestBlockStore, configManager)

	err = trg.Process(context.Background(), slotEvent)

	require.NoError(t, err)
	assert.Empty(t, latestBlockStore.saved)
}

func TestLatestBlockStorePermanentTrigger_Process_SaveError_IsPropagated(t *testing.T) {
	saveErr := errors.New("save failed")
	latestBlockStore := &fakeLatestBlockStore{err: saveErr}
	e := mustSolanaBlockEvent(t)
	configManager := &stubStoreConfigurationManager{
		configs: []store.Configuration{newActiveConfigWithoutEventFeature(t, e.NodeID())},
	}
	trg := NewLatestBlockStorePermanentTrigger[event.SolanaBlockEvent](latestBlockStore, configManager)

	err := trg.Process(context.Background(), e)

	assert.ErrorIs(t, err, saveErr)
}

func TestLatestBlockStorePermanentTrigger_Process_ConfigLoadError_IsPropagated(t *testing.T) {
	loadErr := errors.New("load failed")
	latestBlockStore := &fakeLatestBlockStore{}
	configManager := &stubStoreConfigurationManager{err: loadErr}
	trg := NewLatestBlockStorePermanentTrigger[event.SolanaBlockEvent](latestBlockStore, configManager)

	err := trg.Process(context.Background(), mustSolanaBlockEvent(t))

	assert.ErrorIs(t, err, loadErr)
	assert.Empty(t, latestBlockStore.saved)
}

func TestLatestBlockStorePermanentTrigger_Supports(t *testing.T) {
	trg := NewLatestBlockStorePermanentTrigger[event.SolanaBlockEvent](&fakeLatestBlockStore{}, &stubStoreConfigurationManager{})

	assert.True(t, trg.Supports(mustSolanaBlockEvent(t)))

	slotEvent, err := event.NewSlotEvent(uuid.New(), 1, time.Now())
	require.NoError(t, err)
	assert.False(t, trg.Supports(slotEvent))
}
