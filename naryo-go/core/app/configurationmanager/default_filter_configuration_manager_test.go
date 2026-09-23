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
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter"
)

// --- stubs ---

type stubFilterSourceProvider struct {
	priority int
	items    []descriptor.Filter
	err      error
}

func (s *stubFilterSourceProvider) Load(_ context.Context) ([]descriptor.Filter, error) {
	return s.items, s.err
}

func (s *stubFilterSourceProvider) Priority() int { return s.priority }

type stubFilterDescriptor struct {
	result filter.Filter
	err    error
}

func (s *stubFilterDescriptor) Map() (filter.Filter, error) { return s.result, s.err }

type stubDomainFilter struct{ id uuid.UUID }

func (s *stubDomainFilter) ID() uuid.UUID           { return s.id }
func (s *stubDomainFilter) Name() filter.Name       { return filter.Name("stub") }
func (s *stubDomainFilter) NodeID() uuid.UUID       { return uuid.Nil }
func (s *stubDomainFilter) Type() filter.FilterType { return filter.FilterTypeEvent }

// --- tests ---

func TestDefaultFilterConfigurationManager_Load_NoProviders(t *testing.T) {
	cm := NewDefaultFilterConfigurationManager(nil)

	result, err := cm.Load(context.Background())

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestDefaultFilterConfigurationManager_Load_SingleProvider(t *testing.T) {
	f1 := &stubDomainFilter{id: uuid.New()}
	f2 := &stubDomainFilter{id: uuid.New()}
	provider := &stubFilterSourceProvider{
		priority: 1,
		items: []descriptor.Filter{
			&stubFilterDescriptor{result: f1},
			&stubFilterDescriptor{result: f2},
		},
	}
	cm := NewDefaultFilterConfigurationManager([]sourceprovider.FilterSourceProvider{provider})

	result, err := cm.Load(context.Background())

	require.NoError(t, err)
	assert.Equal(t, []filter.Filter{f1, f2}, result)
}

func TestDefaultFilterConfigurationManager_Load_MultipleProviders_AggregatesAll(t *testing.T) {
	f1 := &stubDomainFilter{id: uuid.New()}
	f2 := &stubDomainFilter{id: uuid.New()}
	providerA := &stubFilterSourceProvider{
		priority: 1,
		items:    []descriptor.Filter{&stubFilterDescriptor{result: f1}},
	}
	providerB := &stubFilterSourceProvider{
		priority: 2,
		items:    []descriptor.Filter{&stubFilterDescriptor{result: f2}},
	}
	cm := NewDefaultFilterConfigurationManager([]sourceprovider.FilterSourceProvider{providerA, providerB})

	result, err := cm.Load(context.Background())

	require.NoError(t, err)
	assert.Equal(t, []filter.Filter{f1, f2}, result)
}

func TestDefaultFilterConfigurationManager_Load_SortsByPriority(t *testing.T) {
	first := &stubDomainFilter{id: uuid.New()}
	second := &stubDomainFilter{id: uuid.New()}
	// Passed in reverse priority order — Load must still yield first before second.
	providerHigh := &stubFilterSourceProvider{
		priority: 2,
		items:    []descriptor.Filter{&stubFilterDescriptor{result: second}},
	}
	providerLow := &stubFilterSourceProvider{
		priority: 1,
		items:    []descriptor.Filter{&stubFilterDescriptor{result: first}},
	}
	cm := NewDefaultFilterConfigurationManager([]sourceprovider.FilterSourceProvider{providerHigh, providerLow})

	result, err := cm.Load(context.Background())

	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, first, result[0])
	assert.Equal(t, second, result[1])
}

func TestDefaultFilterConfigurationManager_Load_ProviderError(t *testing.T) {
	providerErr := errors.New("source unavailable")
	provider := &stubFilterSourceProvider{priority: 1, err: providerErr}
	cm := NewDefaultFilterConfigurationManager([]sourceprovider.FilterSourceProvider{provider})

	_, err := cm.Load(context.Background())

	assert.ErrorIs(t, err, providerErr)
}

func TestDefaultFilterConfigurationManager_Load_DescriptorMapError(t *testing.T) {
	mapErr := errors.New("invalid descriptor")
	provider := &stubFilterSourceProvider{
		priority: 1,
		items:    []descriptor.Filter{&stubFilterDescriptor{err: mapErr}},
	}
	cm := NewDefaultFilterConfigurationManager([]sourceprovider.FilterSourceProvider{provider})

	_, err := cm.Load(context.Background())

	assert.ErrorIs(t, err, mapErr)
}
