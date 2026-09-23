//go:build test

package eventmapper

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

func TestMapTransactionEvent(t *testing.T) {
	nodeID := uuid.New()
	instructions := []event.InstructionData{
		{ProgramID: "Program1", Data: []byte{1, 2, 3}, Accounts: []string{"Account1"}},
	}
	e, err := event.NewSolanaTransactionEvent(nodeID, "TxSig1", 42, instructions, []string{"Account1"}, []string{"log1"}, new("boom"))
	require.NoError(t, err)

	payload := MapTransactionEvent(e)

	assert.Equal(t, nodeID, payload.NodeID)
	assert.Equal(t, "TxSig1", payload.Signature)
	assert.Equal(t, uint64(42), payload.Slot)
	assert.True(t, payload.HasError)
	assert.Nil(t, payload.Reason, "MapTransactionEvent must not decode -- only relay a reason the trigger already computed")
	assert.Equal(t, []string{"Account1"}, payload.Accounts)
	assert.Equal(t, []string{"log1"}, payload.Logs)
	require.Len(t, payload.Instructions, 1)
	assert.Equal(t, "Program1", payload.Instructions[0].ProgramID)
	assert.Equal(t, []byte{1, 2, 3}, payload.Instructions[0].Data)
	assert.Equal(t, []string{"Account1"}, payload.Instructions[0].Accounts)
}

func TestMapTransactionEvent_DecodedReasonSet_PopulatesReason(t *testing.T) {
	e, err := event.NewSolanaTransactionEvent(uuid.New(), "TxSig1", 42, nil, nil, nil, new("boom"))
	require.NoError(t, err)
	e = e.WithDecodedReason("AccountInUse")

	payload := MapTransactionEvent(e)

	require.NotNil(t, payload.Reason)
	assert.Equal(t, "AccountInUse", *payload.Reason)
}
