//go:build test

package event

import (
	"math/big"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
)

func TestSolanaContractEvent_New(t *testing.T) {
	nodeID := uuid.New()
	amount, err := parameter.NewSolanaUintParameter(0, big.NewInt(100), 64)
	assert.NoError(t, err)
	parameters := []parameter.ContractEventParameter{amount}

	t.Run("Valid", func(t *testing.T) {
		evt, err := NewSolanaContractEvent(nodeID, "prog", "sig", 10, parameters, "Transfer", ContractEventStatusConfirmed)
		assert.NoError(t, err)
		assert.Equal(t, TypeContract, evt.EventType())
		assert.Equal(t, nodeID, evt.NodeID())
		assert.Equal(t, parameters, evt.Parameters())
		assert.Equal(t, "prog", evt.ProgramID)
		assert.Equal(t, "sig", evt.Signature)
		assert.Equal(t, uint64(10), evt.Slot)
		assert.Equal(t, "Transfer", evt.EventName())
		assert.Equal(t, ContractEventStatusConfirmed, evt.Status())
	})

	t.Run("NilParametersNormalized", func(t *testing.T) {
		evt, err := NewSolanaContractEvent(nodeID, "prog", "sig", 10, nil, "Transfer", ContractEventStatusConfirmed)
		assert.NoError(t, err)
		assert.Equal(t, []parameter.ContractEventParameter{}, evt.Parameters())
	})

	t.Run("NilNodeID", func(t *testing.T) {
		_, err := NewSolanaContractEvent(uuid.Nil, "prog", "sig", 10, parameters, "Transfer", ContractEventStatusConfirmed)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("EmptyProgramID", func(t *testing.T) {
		_, err := NewSolanaContractEvent(nodeID, "", "sig", 10, parameters, "Transfer", ContractEventStatusConfirmed)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("EmptySignature", func(t *testing.T) {
		_, err := NewSolanaContractEvent(nodeID, "prog", "", 10, parameters, "Transfer", ContractEventStatusConfirmed)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("EmptyEventName", func(t *testing.T) {
		_, err := NewSolanaContractEvent(nodeID, "prog", "sig", 10, parameters, "", ContractEventStatusConfirmed)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("InvalidStatus", func(t *testing.T) {
		_, err := NewSolanaContractEvent(nodeID, "prog", "sig", 10, parameters, "Transfer", ContractEventStatus("BOGUS"))
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}
