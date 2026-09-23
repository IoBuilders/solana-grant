//go:build test

package event

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func TestSolanaTransaction_New(t *testing.T) {
	nodeID := uuid.New()
	instructions := []InstructionData{{ProgramID: "prog", Data: []byte{1, 2}, Accounts: []string{"acc"}}}
	accounts := []string{"acc-1", "acc-2"}
	logs := []string{"log line"}

	t.Run("Valid", func(t *testing.T) {
		tx, err := NewSolanaTransaction(nodeID, "sig", 42, instructions, accounts, logs, nil)
		assert.NoError(t, err)
		assert.Equal(t, nodeID, tx.NodeID())
		assert.Equal(t, "sig", tx.Id())
		assert.Equal(t, "sig", tx.Signature)
		assert.Equal(t, uint64(42), tx.Slot)
		assert.Equal(t, instructions, tx.Instructions)
		assert.Equal(t, accounts, tx.Accounts)
		assert.Equal(t, logs, tx.Logs)
	})

	t.Run("WithError", func(t *testing.T) {
		tx, err := NewSolanaTransaction(nodeID, "sig", 42, instructions, accounts, logs, new("insufficient funds"))
		assert.NoError(t, err)
		assert.True(t, tx.HasError())
	})

	t.Run("NilSlicesNormalized", func(t *testing.T) {
		tx, err := NewSolanaTransaction(nodeID, "sig", 42, nil, nil, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, []InstructionData{}, tx.Instructions)
		assert.Equal(t, []string{}, tx.Accounts)
		assert.Equal(t, []string{}, tx.Logs)
	})

	t.Run("NilNodeID", func(t *testing.T) {
		_, err := NewSolanaTransaction(uuid.Nil, "sig", 42, instructions, accounts, logs, nil)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("EmptySignature", func(t *testing.T) {
		_, err := NewSolanaTransaction(nodeID, "", 42, instructions, accounts, logs, nil)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}

func TestSolanaTransaction_ToSolanaTransactionEvent(t *testing.T) {
	nodeID := uuid.New()
	tx, err := NewSolanaTransaction(nodeID, "sig", 42, nil, nil, nil, nil)
	assert.NoError(t, err)

	txEvent := tx.ToSolanaTransactionEvent()
	assert.Equal(t, TypeTransaction, txEvent.EventType())
	assert.Equal(t, nodeID, txEvent.NodeID())
	assert.Equal(t, "sig", txEvent.Id())
	assert.False(t, txEvent.HasError())
}

func TestSolanaTransactionEvent_New(t *testing.T) {
	nodeID := uuid.New()
	instructions := []InstructionData{{ProgramID: "prog", Data: []byte{1, 2}, Accounts: []string{"acc"}}}
	accounts := []string{"acc-1", "acc-2"}
	logs := []string{"log line"}

	t.Run("Valid", func(t *testing.T) {
		tx, err := NewSolanaTransactionEvent(nodeID, "sig", 42, instructions, accounts, logs, nil)
		assert.NoError(t, err)
		assert.Equal(t, TypeTransaction, tx.EventType())
		assert.Equal(t, nodeID, tx.NodeID())
		assert.Equal(t, "sig", tx.Id())
		assert.Equal(t, "sig", tx.Signature)
		assert.Equal(t, uint64(42), tx.Slot)
		assert.Equal(t, instructions, tx.Instructions)
		assert.Equal(t, accounts, tx.Accounts)
		assert.Equal(t, logs, tx.Logs)
		assert.False(t, tx.HasError())
		assert.Nil(t, tx.DecodedReason)
	})

	t.Run("WithError", func(t *testing.T) {
		tx, err := NewSolanaTransactionEvent(nodeID, "sig", 42, instructions, accounts, logs, new("insufficient funds"))
		assert.NoError(t, err)
		assert.True(t, tx.HasError())
	})

	t.Run("NilSlicesNormalized", func(t *testing.T) {
		tx, err := NewSolanaTransactionEvent(nodeID, "sig", 42, nil, nil, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, []InstructionData{}, tx.Instructions)
		assert.Equal(t, []string{}, tx.Accounts)
		assert.Equal(t, []string{}, tx.Logs)
	})

	t.Run("NilNodeID", func(t *testing.T) {
		_, err := NewSolanaTransactionEvent(uuid.Nil, "sig", 42, instructions, accounts, logs, nil)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("EmptySignature", func(t *testing.T) {
		_, err := NewSolanaTransactionEvent(nodeID, "", 42, instructions, accounts, logs, nil)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}

func TestSolanaTransactionEvent_WithDecodedReason(t *testing.T) {
	tx, err := NewSolanaTransactionEvent(uuid.New(), "sig", 42, nil, nil, nil, nil)
	assert.NoError(t, err)

	withReason := tx.WithDecodedReason("InsufficientFunds")

	assert.Nil(t, tx.DecodedReason, "the receiver must not be mutated")
	assert.NotNil(t, withReason.DecodedReason)
	assert.Equal(t, "InsufficientFunds", *withReason.DecodedReason)
}
