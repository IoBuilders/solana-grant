//go:build test

package feature

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func newTarget(t *testing.T, targetType TargetType, destination string) EventTarget {
	t.Helper()
	target, err := NewEventTarget(targetType, Destination(destination))
	require.NoError(t, err)
	return target
}

func TestEventTarget_New(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		target, err := NewEventTarget(TargetTypeTransaction, Destination("transactions"))
		assert.NoError(t, err)
		assert.Equal(t, TargetTypeTransaction, target.TargetType())
		assert.Equal(t, Destination("transactions"), target.Destination())
	})

	t.Run("UnknownTargetType", func(t *testing.T) {
		_, err := NewEventTarget(TargetType("SLOT"), Destination("slots"))
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	// A SQL store cannot honour a destination, so requiring one would be
	// requiring something unusable. The backends that need it — HTTP posts to
	// it — gate it themselves with NewDestination.
	t.Run("DestinationIsOptional", func(t *testing.T) {
		target, err := NewEventTarget(TargetTypeBlock, Destination(""))
		require.NoError(t, err)
		assert.Equal(t, TargetTypeBlock, target.TargetType())
		assert.Empty(t, target.Destination())
	})
}

func TestEventConfiguration_New(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		targets := []EventTarget{
			newTarget(t, TargetTypeBlock, "blocks"),
			newTarget(t, TargetTypeTransaction, "transactions"),
		}

		cfg, err := NewEventConfiguration(StrategyBlockBased, targets)
		assert.NoError(t, err)
		assert.Equal(t, TypeEvent, cfg.Type())
		assert.Equal(t, StrategyBlockBased, cfg.Strategy())
		assert.Len(t, cfg.Targets(), 2)
	})

	// Unlike the Java parent, a partial set of targets is accepted.
	t.Run("SubsetOfTargetsIsAccepted", func(t *testing.T) {
		cfg, err := NewEventConfiguration(StrategyBlockBased, []EventTarget{
			newTarget(t, TargetTypeTransaction, "transactions"),
		})
		require.NoError(t, err)

		_, hasTransaction := cfg.Target(TargetTypeTransaction)
		assert.True(t, hasTransaction)
		_, hasBlock := cfg.Target(TargetTypeBlock)
		assert.False(t, hasBlock)
	})

	t.Run("DuplicateTargetType", func(t *testing.T) {
		_, err := NewEventConfiguration(StrategyBlockBased, []EventTarget{
			newTarget(t, TargetTypeTransaction, "transactions"),
			newTarget(t, TargetTypeTransaction, "other-transactions"),
		})
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("NoTargets", func(t *testing.T) {
		_, err := NewEventConfiguration(StrategyBlockBased, nil)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("UnknownStrategy", func(t *testing.T) {
		_, err := NewEventConfiguration(Strategy("SLOT_BASED"), []EventTarget{
			newTarget(t, TargetTypeTransaction, "transactions"),
		})
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}

func TestFilterSyncConfiguration_New(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		cfg, err := NewFilterSyncConfiguration(Destination("filters"))
		assert.NoError(t, err)
		assert.Equal(t, TypeFilterSync, cfg.Type())
		assert.Equal(t, Destination("filters"), cfg.Destination())
	})

	t.Run("DestinationIsOptional", func(t *testing.T) {
		cfg, err := NewFilterSyncConfiguration(Destination(""))
		require.NoError(t, err)
		assert.Equal(t, TypeFilterSync, cfg.Type())
		assert.Empty(t, cfg.Destination())
	})
}

func TestTypes_IsValid(t *testing.T) {
	assert.True(t, TypeEvent.IsValid())
	assert.True(t, TypeFilterSync.IsValid())
	assert.False(t, Type("SLOT").IsValid())

	assert.True(t, StrategyBlockBased.IsValid())
	assert.False(t, Strategy("").IsValid())

	assert.True(t, TargetTypeBlock.IsValid())
	assert.True(t, TargetTypeTransaction.IsValid())
	assert.True(t, TargetTypeContractEvent.IsValid())
	assert.False(t, TargetType("FILTER").IsValid())
}
