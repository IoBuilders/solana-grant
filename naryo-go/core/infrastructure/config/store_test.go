//go:build test

package config

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/store"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/store/feature"
)

const testNodeID = "00000000-0000-0000-0000-000000000001"

func fullStoreProperties() *StoreProperties {
	return &StoreProperties{
		NodeID: testNodeID,
		Type:   "GORM",
		Features: []*StoreFeatureProperties{
			{
				Type:     "EVENT",
				Strategy: "BLOCK_BASED",
				Targets: []*StoreTargetProperties{
					{Type: "TRANSACTION", Destination: "transactions"},
				},
			},
			{Type: "FILTER_SYNC", Destination: "filters"},
			{Type: "LATEST_BLOCK", Destination: "latest_blocks"},
		},
	}
}

func TestStoreProperties_Map_Active(t *testing.T) {
	configuration, err := fullStoreProperties().Map()
	require.NoError(t, err)

	assert.Equal(t, uuid.MustParse(testNodeID), configuration.NodeID())
	assert.Equal(t, store.StateActive, configuration.State())

	active, ok := configuration.(*store.ActiveConfiguration)
	require.True(t, ok)
	assert.Equal(t, store.TypeGorm, active.Type())

	eventFeature, found := active.Feature(feature.TypeEvent)
	require.True(t, found)
	eventConfiguration, ok := eventFeature.(*feature.EventConfiguration)
	require.True(t, ok)
	assert.Equal(t, feature.StrategyBlockBased, eventConfiguration.Strategy())

	target, found := eventConfiguration.Target(feature.TargetTypeTransaction)
	require.True(t, found)
	assert.Equal(t, feature.Destination("transactions"), target.Destination())

	filterFeature, found := active.Feature(feature.TypeFilterSync)
	require.True(t, found)
	filterConfiguration, ok := filterFeature.(*feature.FilterSyncConfiguration)
	require.True(t, ok)
	assert.Equal(t, feature.Destination("filters"), filterConfiguration.Destination())

	latestBlockFeature, found := active.Feature(feature.TypeLatestBlock)
	require.True(t, found)
	latestBlockFeatureConfiguration, ok := latestBlockFeature.(*feature.LatestBlockConfiguration)
	require.True(t, ok)
	assert.Equal(t, feature.Destination("latest_blocks"), latestBlockFeatureConfiguration.Destination())
}

// The file may use either case; the domain constants are upper case.
func TestStoreProperties_Map_IsCaseInsensitive(t *testing.T) {
	properties := &StoreProperties{
		NodeID: testNodeID,
		Type:   "gorm",
		Features: []*StoreFeatureProperties{
			{
				Type:     "event",
				Strategy: "block_based",
				Targets:  []*StoreTargetProperties{{Type: "transaction", Destination: "transactions"}},
			},
		},
	}

	configuration, err := properties.Map()
	require.NoError(t, err)
	active, ok := configuration.(*store.ActiveConfiguration)
	require.True(t, ok)
	assert.Equal(t, store.TypeGorm, active.Type())
}

// BLOCK_BASED is the only strategy, so leaving it out is not an error.
func TestStoreProperties_Map_DefaultsStrategy(t *testing.T) {
	properties := &StoreProperties{
		NodeID: testNodeID,
		Type:   "GORM",
		Features: []*StoreFeatureProperties{
			{Type: "EVENT", Targets: []*StoreTargetProperties{{Type: "TRANSACTION", Destination: "transactions"}}},
		},
	}

	configuration, err := properties.Map()
	require.NoError(t, err)

	active := configuration.(*store.ActiveConfiguration)
	eventFeature, _ := active.Feature(feature.TypeEvent)
	assert.Equal(t, feature.StrategyBlockBased, eventFeature.(*feature.EventConfiguration).Strategy())
}

func TestStoreProperties_Map_Inactive(t *testing.T) {
	// An inactive entry persists nothing, so it needs no backend.
	properties := &StoreProperties{NodeID: testNodeID, State: "INACTIVE"}

	configuration, err := properties.Map()
	require.NoError(t, err)
	assert.Equal(t, store.StateInactive, configuration.State())
	assert.Equal(t, uuid.MustParse(testNodeID), configuration.NodeID())
}

func TestStoreProperties_Map_Errors(t *testing.T) {
	t.Run("InvalidNodeID", func(t *testing.T) {
		properties := fullStoreProperties()
		properties.NodeID = "not-a-uuid"
		_, err := properties.Map()
		assert.ErrorContains(t, err, "invalid nodeId")
	})

	t.Run("UnknownState", func(t *testing.T) {
		properties := fullStoreProperties()
		properties.State = "PAUSED"
		_, err := properties.Map()
		assert.ErrorContains(t, err, "unsupported state")
	})

	// A backend nothing implements must be reported on load, not at the first
	// save when no Store claims it.
	t.Run("UnknownBackend", func(t *testing.T) {
		properties := fullStoreProperties()
		properties.Type = "MONGO"
		_, err := properties.Map()
		assert.ErrorContains(t, err, "unknown store type")
	})

	t.Run("MissingType", func(t *testing.T) {
		properties := fullStoreProperties()
		properties.Type = ""
		_, err := properties.Map()
		assert.ErrorContains(t, err, "unknown store type")
	})

	t.Run("UnknownFeatureType", func(t *testing.T) {
		properties := fullStoreProperties()
		properties.Features = []*StoreFeatureProperties{{Type: "SLOT_SYNC"}}
		_, err := properties.Map()
		assert.ErrorContains(t, err, "unsupported feature type")
	})

	t.Run("DuplicateFeature", func(t *testing.T) {
		properties := fullStoreProperties()
		properties.Features = []*StoreFeatureProperties{
			{Type: "FILTER_SYNC", Destination: "a"},
			{Type: "FILTER_SYNC", Destination: "b"},
		}
		_, err := properties.Map()
		assert.ErrorContains(t, err, "duplicate feature")
	})

	t.Run("UnknownTargetType", func(t *testing.T) {
		properties := fullStoreProperties()
		properties.Features = []*StoreFeatureProperties{
			{Type: "EVENT", Targets: []*StoreTargetProperties{{Type: "SLOT", Destination: "slots"}}},
		}
		_, err := properties.Map()
		assert.Error(t, err)
	})

}

// A destination only means something to a backend that can route by it. The SQL
// adapter writes to the table its model names, so the file is not made to
// declare one it cannot use.
func TestStoreProperties_Map_DestinationIsOptional(t *testing.T) {
	properties := fullStoreProperties()
	properties.Features = []*StoreFeatureProperties{
		{Type: "EVENT", Targets: []*StoreTargetProperties{{Type: "TRANSACTION"}}},
		{Type: "FILTER_SYNC"},
		{Type: "LATEST_BLOCK"},
	}

	configuration, err := properties.Map()
	require.NoError(t, err)

	active, ok := configuration.(*store.ActiveConfiguration)
	require.True(t, ok)

	eventFeature, found := active.Feature(feature.TypeEvent)
	require.True(t, found)
	target, found := eventFeature.(*feature.EventConfiguration).Target(feature.TargetTypeTransaction)
	require.True(t, found)
	assert.Empty(t, target.Destination())

	filterFeature, found := active.Feature(feature.TypeFilterSync)
	require.True(t, found)
	assert.Empty(t, filterFeature.(*feature.FilterSyncConfiguration).Destination())

	latestBlockFeature, found := active.Feature(feature.TypeLatestBlock)
	require.True(t, found)
	assert.Empty(t, latestBlockFeature.(*feature.LatestBlockConfiguration).Destination())
}

func TestEnvStoreSourceProvider_Load(t *testing.T) {
	properties := &EnvironmentProperties{Stores: []*StoreProperties{
		fullStoreProperties(),
		nil, // a hole in the list must be skipped, not panic
	}}

	provider := NewEnvStoreSourceProvider(properties)
	assert.Equal(t, 1, provider.Priority())

	descriptors, err := provider.Load(context.Background())
	require.NoError(t, err)
	require.Len(t, descriptors, 1)

	configuration, err := descriptors[0].Map()
	require.NoError(t, err)
	assert.Equal(t, store.StateActive, configuration.State())
}

func TestEnvStoreSourceProvider_Load_NoStores(t *testing.T) {
	provider := NewEnvStoreSourceProvider(&EnvironmentProperties{})

	descriptors, err := provider.Load(context.Background())
	require.NoError(t, err)
	assert.Empty(t, descriptors)
}
