//go:build test

package filter

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func TestNewSyncState(t *testing.T) {
	filterID := uuid.New()

	t.Run("Valid", func(t *testing.T) {
		state, err := NewSyncState(filterID, 100, SyncStatusSyncing)
		assert.NoError(t, err)
		assert.Equal(t, filterID, state.FilterID)
		assert.Equal(t, uint64(100), state.LastProcessedSlot)
		assert.Equal(t, SyncStatusSyncing, state.Status)
	})

	t.Run("NilFilterID", func(t *testing.T) {
		_, err := NewSyncState(uuid.Nil, 0, SyncStatusIdle)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("InvalidStatus", func(t *testing.T) {
		_, err := NewSyncState(filterID, 0, SyncStatus("PAUSED"))
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}
