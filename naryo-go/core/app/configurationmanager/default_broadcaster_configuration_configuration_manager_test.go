//go:build test

package configurationmanager

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	configurationmapperregistry "gitlab.com/iobuilders/projects/eng/naryo-go/core/app/configurationmapper"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/sourceprovider"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
)

// --- stubs ---

type stubBroadcasterConfigurationSourceProvider struct {
	priority int
	items    []descriptor.BroadcasterConfiguration
	err      error
}

func (s *stubBroadcasterConfigurationSourceProvider) Load(_ context.Context) ([]descriptor.BroadcasterConfiguration, error) {
	return s.items, s.err
}

func (s *stubBroadcasterConfigurationSourceProvider) Priority() int { return s.priority }

type stubBroadcasterConfigurationDescriptor struct {
	result broadcaster.Configuration
	err    error
}

func (s *stubBroadcasterConfigurationDescriptor) Map() (broadcaster.Configuration, error) {
	return s.result, s.err
}

type stubDomainBroadcasterConfiguration struct{ id uuid.UUID }

func (s *stubDomainBroadcasterConfiguration) ID() uuid.UUID          { return s.id }
func (s *stubDomainBroadcasterConfiguration) Type() broadcaster.Type { return broadcaster.TypeHTTP }
func (s *stubDomainBroadcasterConfiguration) Validate() error        { return nil }
func (s *stubDomainBroadcasterConfiguration) AdditionalProperties() map[string]interface{} {
	return nil
}

// identityMapperRegistry registers a passthrough mapper for broadcaster.TypeHTTP, the type
// every stubDomainBroadcasterConfiguration reports, so Map() returns the source unchanged.
func identityMapperRegistry(t *testing.T) configurationmapperregistry.ConfigurationMapperRegistry[broadcaster.Configuration] {
	registry := configurationmapperregistry.NewBaseConfigurationMapperRegistry[broadcaster.Configuration]()
	err := registry.Register(context.Background(), broadcaster.TypeHTTP.String(), func(_ context.Context, source broadcaster.Configuration) (broadcaster.Configuration, error) {
		return source, nil
	})
	require.NoError(t, err)
	return registry
}

// --- tests ---

func TestDefaultBroadcasterConfigurationConfigurationManager_Load_NoProviders(t *testing.T) {
	cm := NewDefaultBroadcasterConfigurationConfigurationManager(nil, identityMapperRegistry(t))

	result, err := cm.Load(context.Background())

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestDefaultBroadcasterConfigurationConfigurationManager_Load_SingleProvider(t *testing.T) {
	cfg1 := &stubDomainBroadcasterConfiguration{id: uuid.New()}
	cfg2 := &stubDomainBroadcasterConfiguration{id: uuid.New()}
	provider := &stubBroadcasterConfigurationSourceProvider{
		priority: 1,
		items: []descriptor.BroadcasterConfiguration{
			&stubBroadcasterConfigurationDescriptor{result: cfg1},
			&stubBroadcasterConfigurationDescriptor{result: cfg2},
		},
	}
	cm := NewDefaultBroadcasterConfigurationConfigurationManager([]sourceprovider.BroadcasterConfigurationSourceProvider{provider}, identityMapperRegistry(t))

	result, err := cm.Load(context.Background())

	require.NoError(t, err)
	assert.Equal(t, []broadcaster.Configuration{cfg1, cfg2}, result)
}

func TestDefaultBroadcasterConfigurationConfigurationManager_Load_MultipleProviders_AggregatesAll(t *testing.T) {
	cfg1 := &stubDomainBroadcasterConfiguration{id: uuid.New()}
	cfg2 := &stubDomainBroadcasterConfiguration{id: uuid.New()}
	providerA := &stubBroadcasterConfigurationSourceProvider{
		priority: 1,
		items:    []descriptor.BroadcasterConfiguration{&stubBroadcasterConfigurationDescriptor{result: cfg1}},
	}
	providerB := &stubBroadcasterConfigurationSourceProvider{
		priority: 2,
		items:    []descriptor.BroadcasterConfiguration{&stubBroadcasterConfigurationDescriptor{result: cfg2}},
	}
	cm := NewDefaultBroadcasterConfigurationConfigurationManager([]sourceprovider.BroadcasterConfigurationSourceProvider{providerA, providerB}, identityMapperRegistry(t))

	result, err := cm.Load(context.Background())

	require.NoError(t, err)
	assert.Equal(t, []broadcaster.Configuration{cfg1, cfg2}, result)
}

func TestDefaultBroadcasterConfigurationConfigurationManager_Load_SortsByPriority(t *testing.T) {
	first := &stubDomainBroadcasterConfiguration{id: uuid.New()}
	second := &stubDomainBroadcasterConfiguration{id: uuid.New()}
	// Passed in reverse priority order — Load must still yield first before second.
	providerHigh := &stubBroadcasterConfigurationSourceProvider{
		priority: 2,
		items:    []descriptor.BroadcasterConfiguration{&stubBroadcasterConfigurationDescriptor{result: second}},
	}
	providerLow := &stubBroadcasterConfigurationSourceProvider{
		priority: 1,
		items:    []descriptor.BroadcasterConfiguration{&stubBroadcasterConfigurationDescriptor{result: first}},
	}
	cm := NewDefaultBroadcasterConfigurationConfigurationManager([]sourceprovider.BroadcasterConfigurationSourceProvider{providerHigh, providerLow}, identityMapperRegistry(t))

	result, err := cm.Load(context.Background())

	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, first, result[0])
	assert.Equal(t, second, result[1])
}

func TestDefaultBroadcasterConfigurationConfigurationManager_Load_ProviderError(t *testing.T) {
	providerErr := errors.New("source unavailable")
	provider := &stubBroadcasterConfigurationSourceProvider{priority: 1, err: providerErr}
	cm := NewDefaultBroadcasterConfigurationConfigurationManager([]sourceprovider.BroadcasterConfigurationSourceProvider{provider}, identityMapperRegistry(t))

	_, err := cm.Load(context.Background())

	assert.ErrorIs(t, err, providerErr)
}

func TestDefaultBroadcasterConfigurationConfigurationManager_Load_DescriptorMapError(t *testing.T) {
	mapErr := errors.New("invalid descriptor")
	provider := &stubBroadcasterConfigurationSourceProvider{
		priority: 1,
		items:    []descriptor.BroadcasterConfiguration{&stubBroadcasterConfigurationDescriptor{err: mapErr}},
	}
	cm := NewDefaultBroadcasterConfigurationConfigurationManager([]sourceprovider.BroadcasterConfigurationSourceProvider{provider}, identityMapperRegistry(t))

	_, err := cm.Load(context.Background())

	assert.ErrorIs(t, err, mapErr)
}
