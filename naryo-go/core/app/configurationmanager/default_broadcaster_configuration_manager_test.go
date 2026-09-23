//go:build test

package configurationmanager

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/sourceprovider"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster/target"
)

// --- stubs ---

type stubBroadcasterSourceProvider struct {
	priority int
	items    []descriptor.Broadcaster
	err      error
}

func (s *stubBroadcasterSourceProvider) Load(_ context.Context) ([]descriptor.Broadcaster, error) {
	return s.items, s.err
}

func (s *stubBroadcasterSourceProvider) Priority() int { return s.priority }

type stubBroadcasterDescriptor struct {
	result *broadcaster.Broadcaster
	err    error
}

func (s *stubBroadcasterDescriptor) Map() (*broadcaster.Broadcaster, error) {
	return s.result, s.err
}

func newBroadcaster(t *testing.T) *broadcaster.Broadcaster {
	t.Helper()
	dst := target.Destination("/hook")
	trg, err := target.NewBlockTarget([]target.Destination{dst})
	require.NoError(t, err)
	b, err := broadcaster.NewBroadcaster(uuid.New(), trg, uuid.New())
	require.NoError(t, err)
	return b
}

// --- tests ---

func TestDefaultBroadcasterConfigurationManager_Load_NoProviders(t *testing.T) {
	cm := NewDefaultBroadcasterConfigurationManager(nil)

	result, err := cm.Load(context.Background())

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestDefaultBroadcasterConfigurationManager_Load_SingleProvider(t *testing.T) {
	b1 := newBroadcaster(t)
	b2 := newBroadcaster(t)
	provider := &stubBroadcasterSourceProvider{
		priority: 1,
		items: []descriptor.Broadcaster{
			&stubBroadcasterDescriptor{result: b1},
			&stubBroadcasterDescriptor{result: b2},
		},
	}
	cm := NewDefaultBroadcasterConfigurationManager([]sourceprovider.BroadcasterSourceProvider{provider})

	result, err := cm.Load(context.Background())

	require.NoError(t, err)
	assert.Equal(t, []*broadcaster.Broadcaster{b1, b2}, result)
}

func TestDefaultBroadcasterConfigurationManager_Load_MultipleProviders_AggregatesAll(t *testing.T) {
	b1 := newBroadcaster(t)
	b2 := newBroadcaster(t)
	providerA := &stubBroadcasterSourceProvider{
		priority: 1,
		items:    []descriptor.Broadcaster{&stubBroadcasterDescriptor{result: b1}},
	}
	providerB := &stubBroadcasterSourceProvider{
		priority: 2,
		items:    []descriptor.Broadcaster{&stubBroadcasterDescriptor{result: b2}},
	}
	cm := NewDefaultBroadcasterConfigurationManager([]sourceprovider.BroadcasterSourceProvider{providerA, providerB})

	result, err := cm.Load(context.Background())

	require.NoError(t, err)
	assert.Equal(t, []*broadcaster.Broadcaster{b1, b2}, result)
}

func TestDefaultBroadcasterConfigurationManager_Load_SortsByPriority(t *testing.T) {
	first := newBroadcaster(t)
	second := newBroadcaster(t)
	// Passed in reverse priority order — Load must still yield first before second.
	providerHigh := &stubBroadcasterSourceProvider{
		priority: 2,
		items:    []descriptor.Broadcaster{&stubBroadcasterDescriptor{result: second}},
	}
	providerLow := &stubBroadcasterSourceProvider{
		priority: 1,
		items:    []descriptor.Broadcaster{&stubBroadcasterDescriptor{result: first}},
	}
	cm := NewDefaultBroadcasterConfigurationManager([]sourceprovider.BroadcasterSourceProvider{providerHigh, providerLow})

	result, err := cm.Load(context.Background())

	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, first, result[0])
	assert.Equal(t, second, result[1])
}

func TestDefaultBroadcasterConfigurationManager_Load_ProviderError(t *testing.T) {
	providerErr := errors.New("source unavailable")
	provider := &stubBroadcasterSourceProvider{priority: 1, err: providerErr}
	cm := NewDefaultBroadcasterConfigurationManager([]sourceprovider.BroadcasterSourceProvider{provider})

	_, err := cm.Load(context.Background())

	assert.ErrorIs(t, err, providerErr)
}

func TestDefaultBroadcasterConfigurationManager_Load_DescriptorMapError(t *testing.T) {
	mapErr := errors.New("invalid descriptor")
	provider := &stubBroadcasterSourceProvider{
		priority: 1,
		items:    []descriptor.Broadcaster{&stubBroadcasterDescriptor{err: mapErr}},
	}
	cm := NewDefaultBroadcasterConfigurationManager([]sourceprovider.BroadcasterSourceProvider{provider})

	_, err := cm.Load(context.Background())

	assert.ErrorIs(t, err, mapErr)
}
