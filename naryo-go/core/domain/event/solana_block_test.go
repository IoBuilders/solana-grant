//go:build test

package event

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func TestSolanaBlockEvent_New(t *testing.T) {
	nodeID := uuid.New()
	blockTime := int64(1_700_000_000)

	t.Run("Valid", func(t *testing.T) {
		tx, err := NewSolanaTransaction(nodeID, "sig", 10, nil, nil, nil, nil)
		assert.NoError(t, err)

		block, err := NewSolanaBlockEvent(nodeID, 10, "hash", &blockTime, []SolanaTransaction{tx})
		assert.NoError(t, err)
		assert.Equal(t, TypeBlock, block.EventType())
		assert.Equal(t, nodeID, block.NodeID())
		assert.Equal(t, uint64(10), block.Slot)
		assert.Equal(t, "hash", block.Blockhash)
		assert.Equal(t, &blockTime, block.BlockTime)
		assert.Equal(t, []BlockTransaction{tx}, block.Transactions())
	})

	t.Run("NilTransactionsNormalized", func(t *testing.T) {
		block, err := NewSolanaBlockEvent(nodeID, 10, "hash", nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, []BlockTransaction{}, block.Transactions())
	})

	t.Run("NilNodeID", func(t *testing.T) {
		_, err := NewSolanaBlockEvent(uuid.Nil, 10, "hash", nil, nil)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("EmptyBlockhash", func(t *testing.T) {
		_, err := NewSolanaBlockEvent(nodeID, 10, "", nil, nil)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}
