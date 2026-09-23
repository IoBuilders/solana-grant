//go:build test

package broadcaster

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster/target"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func newValidTarget(t *testing.T) target.Target {
	t.Helper()
	destination, err := target.NewDestination("/webhook")
	assert.NoError(t, err)
	blockTarget, err := target.NewBlockTarget([]target.Destination{destination})
	assert.NoError(t, err)
	return blockTarget
}

func TestBroadcaster_New(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		id, configurationID := uuid.New(), uuid.New()
		tgt := newValidTarget(t)

		b, err := NewBroadcaster(id, tgt, configurationID)
		assert.NoError(t, err)
		assert.Equal(t, id, b.ID)
		assert.Equal(t, tgt, b.Target)
		assert.Equal(t, configurationID, b.ConfigurationID)
	})

	t.Run("NilID", func(t *testing.T) {
		_, err := NewBroadcaster(uuid.Nil, newValidTarget(t), uuid.New())
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("NilTarget", func(t *testing.T) {
		_, err := NewBroadcaster(uuid.New(), nil, uuid.New())
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("InvalidTarget", func(t *testing.T) {
		_, err := NewBroadcaster(uuid.New(), target.BlockTarget{}, uuid.New())
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("NilConfigurationID", func(t *testing.T) {
		_, err := NewBroadcaster(uuid.New(), newValidTarget(t), uuid.Nil)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}
