//go:build test

package store

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/store/feature"
)

func newEventFeature(t *testing.T) feature.Configuration {
	t.Helper()
	target, err := feature.NewEventTarget(feature.TargetTypeTransaction, feature.Destination("transactions"))
	require.NoError(t, err)
	cfg, err := feature.NewEventConfiguration(feature.StrategyBlockBased, []feature.EventTarget{target})
	require.NoError(t, err)
	return cfg
}

func TestInactiveConfiguration_New(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		nodeID := uuid.New()
		cfg, err := NewInactiveConfiguration(nodeID)
		assert.NoError(t, err)
		assert.Equal(t, nodeID, cfg.NodeID())
		assert.Equal(t, StateInactive, cfg.State())
	})

	t.Run("NilNodeID", func(t *testing.T) {
		_, err := NewInactiveConfiguration(uuid.Nil)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}

func TestActiveConfiguration_New(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		nodeID := uuid.New()
		eventFeature := newEventFeature(t)

		cfg, err := NewActiveConfiguration(nodeID, TypeGorm, map[feature.Type]feature.Configuration{
			feature.TypeEvent: eventFeature,
		})
		require.NoError(t, err)
		assert.Equal(t, nodeID, cfg.NodeID())
		assert.Equal(t, StateActive, cfg.State())
		assert.Equal(t, TypeGorm, cfg.Type())

		got, found := cfg.Feature(feature.TypeEvent)
		assert.True(t, found)
		assert.Equal(t, eventFeature, got)

		_, found = cfg.Feature(feature.TypeFilterSync)
		assert.False(t, found)
	})

	t.Run("NilNodeID", func(t *testing.T) {
		_, err := NewActiveConfiguration(uuid.Nil, TypeGorm, map[feature.Type]feature.Configuration{
			feature.TypeEvent: newEventFeature(t),
		})
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	// An unknown backend must fail on load rather than at the first save, when
	// no Store would claim the configuration through Supports.
	t.Run("UnknownStoreType", func(t *testing.T) {
		_, err := NewActiveConfiguration(uuid.New(), Type("MONGO"), map[feature.Type]feature.Configuration{
			feature.TypeEvent: newEventFeature(t),
		})
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("NoFeatures", func(t *testing.T) {
		_, err := NewActiveConfiguration(uuid.New(), TypeGorm, nil)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("NilFeature", func(t *testing.T) {
		_, err := NewActiveConfiguration(uuid.New(), TypeGorm, map[feature.Type]feature.Configuration{
			feature.TypeEvent: nil,
		})
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	// A feature filed under the wrong key would make Feature() return
	// something the caller then type-asserts to the wrong shape.
	t.Run("FeatureRegisteredUnderMismatchedKey", func(t *testing.T) {
		_, err := NewActiveConfiguration(uuid.New(), TypeGorm, map[feature.Type]feature.Configuration{
			feature.TypeFilterSync: newEventFeature(t),
		})
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("UnknownFeatureType", func(t *testing.T) {
		_, err := NewActiveConfiguration(uuid.New(), TypeGorm, map[feature.Type]feature.Configuration{
			feature.Type("SLOT_SYNC"): newEventFeature(t),
		})
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}

func TestStateAndType_IsValid(t *testing.T) {
	assert.True(t, StateActive.IsValid())
	assert.True(t, StateInactive.IsValid())
	assert.False(t, State("PAUSED").IsValid())

	assert.True(t, TypeGorm.IsValid())
	assert.False(t, Type("MONGO").IsValid())
}
