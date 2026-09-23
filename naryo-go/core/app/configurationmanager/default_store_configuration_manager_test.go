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
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/store"
)

// --- stubs ---

type stubStoreSourceProvider struct {
	priority int
	items    []descriptor.Store
	err      error
}

func (s *stubStoreSourceProvider) Load(_ context.Context) ([]descriptor.Store, error) {
	return s.items, s.err
}

func (s *stubStoreSourceProvider) Priority() int { return s.priority }

type stubStoreDescriptor struct {
	result store.Configuration
	err    error
}

func (s *stubStoreDescriptor) Map() (store.Configuration, error) {
	return s.result, s.err
}

func newInactiveConfiguration(t *testing.T) store.Configuration {
	t.Helper()
	configuration, err := store.NewInactiveConfiguration(uuid.New())
	require.NoError(t, err)
	return configuration
}

// --- tests ---

func TestDefaultStoreConfigurationManager_Load_NoProviders(t *testing.T) {
	manager := NewDefaultStoreConfigurationManager(nil)

	configurations, err := manager.Load(context.Background())
	require.NoError(t, err)
	assert.Empty(t, configurations)
}

func TestDefaultStoreConfigurationManager_Load_MapsDescriptors(t *testing.T) {
	first := newInactiveConfiguration(t)
	second := newInactiveConfiguration(t)

	manager := NewDefaultStoreConfigurationManager([]sourceprovider.StoreSourceProvider{
		&stubStoreSourceProvider{priority: 1, items: []descriptor.Store{
			&stubStoreDescriptor{result: first},
			&stubStoreDescriptor{result: second},
		}},
	})

	configurations, err := manager.Load(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []store.Configuration{first, second}, configurations)
}

// Providers are consulted in ascending priority order.
func TestDefaultStoreConfigurationManager_Load_RespectsPriority(t *testing.T) {
	low := newInactiveConfiguration(t)
	high := newInactiveConfiguration(t)

	manager := NewDefaultStoreConfigurationManager([]sourceprovider.StoreSourceProvider{
		&stubStoreSourceProvider{priority: 10, items: []descriptor.Store{&stubStoreDescriptor{result: low}}},
		&stubStoreSourceProvider{priority: 1, items: []descriptor.Store{&stubStoreDescriptor{result: high}}},
	})

	configurations, err := manager.Load(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []store.Configuration{high, low}, configurations)
}

func TestDefaultStoreConfigurationManager_Load_ProviderError(t *testing.T) {
	sentinel := errors.New("provider down")
	manager := NewDefaultStoreConfigurationManager([]sourceprovider.StoreSourceProvider{
		&stubStoreSourceProvider{priority: 1, err: sentinel},
	})

	_, err := manager.Load(context.Background())
	assert.ErrorIs(t, err, sentinel)
}

func TestDefaultStoreConfigurationManager_Load_DescriptorError(t *testing.T) {
	sentinel := errors.New("bad entry")
	manager := NewDefaultStoreConfigurationManager([]sourceprovider.StoreSourceProvider{
		&stubStoreSourceProvider{priority: 1, items: []descriptor.Store{
			&stubStoreDescriptor{err: sentinel},
		}},
	})

	_, err := manager.Load(context.Background())
	assert.ErrorIs(t, err, sentinel)
}
