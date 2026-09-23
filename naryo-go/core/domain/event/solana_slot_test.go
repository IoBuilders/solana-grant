//go:build test

package event

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func TestSlotEvent_New(t *testing.T) {
	nodeID := uuid.New()
	receivedAt := time.Now()

	t.Run("Valid", func(t *testing.T) {
		slot, err := NewSlotEvent(nodeID, 42, receivedAt)
		assert.NoError(t, err)
		assert.Equal(t, TypeSlot, slot.EventType())
		assert.Equal(t, nodeID, slot.NodeID())
		assert.Equal(t, uint64(42), slot.Slot)
		assert.Equal(t, receivedAt, slot.ReceivedAt)
	})

	t.Run("NilNodeID", func(t *testing.T) {
		_, err := NewSlotEvent(uuid.Nil, 42, receivedAt)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}
