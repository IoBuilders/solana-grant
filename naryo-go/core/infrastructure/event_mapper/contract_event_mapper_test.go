//go:build test

package eventmapper

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
)

func TestMapContractEvent(t *testing.T) {
	programId := "programId"
	signature := "signature"
	eventName := "eventName"
	nodeID := uuid.New()
	filterID := uuid.New()
	param, err := parameter.NewSolanaStringParameter(0, "value")
	require.NoError(t, err)
	e, err := event.NewSolanaContractEvent(nodeID, programId, signature, 42, []parameter.ContractEventParameter{param}, eventName, event.ContractEventStatusConfirmed)
	require.NoError(t, err)

	payload := MapContractEvent(e, &filterID)

	assert.Equal(t, eventName, payload.EventName)
	assert.Equal(t, nodeID, payload.NodeID)
	require.NotNil(t, payload.FilterID)
	assert.Equal(t, filterID, *payload.FilterID)
	assert.Equal(t, programId, payload.ProgramID)
	assert.Equal(t, signature, payload.Signature)
	assert.Equal(t, uint64(42), payload.Slot)
	require.Len(t, payload.Parameters, 1)
	assert.Equal(t, param.Position(), payload.Parameters[0].Position)
	assert.Equal(t, "STRING", payload.Parameters[0].Type)
	assert.Equal(t, param.Value(), payload.Parameters[0].Value)
}

func TestMapContractEvent_NoFilter(t *testing.T) {
	param, err := parameter.NewSolanaStringParameter(0, "value")
	require.NoError(t, err)
	e, err := event.NewSolanaContractEvent(uuid.New(), "Program1", "Sig1", 42, []parameter.ContractEventParameter{param}, "Transfer", event.ContractEventStatusConfirmed)
	require.NoError(t, err)

	payload := MapContractEvent(e, nil)

	assert.Nil(t, payload.FilterID)
}
