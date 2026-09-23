//go:build test

package eventmapper

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

func TestMapBlockEvent(t *testing.T) {
	nodeID := uuid.New()
	tx, err := event.NewSolanaTransaction(nodeID, "TxSig1", 42, nil, nil, nil, nil)
	require.NoError(t, err)
	blockTime := int64(1737400000)
	e, err := event.NewSolanaBlockEvent(nodeID, 42, "Blockhash1", &blockTime, []event.SolanaTransaction{tx})
	require.NoError(t, err)

	payload := MapBlockEvent(e)

	assert.Equal(t, nodeID, payload.NodeID)
	assert.Equal(t, uint64(42), payload.Slot)
	assert.Equal(t, "Blockhash1", payload.Blockhash)
	require.NotNil(t, payload.BlockTime)
	assert.Equal(t, blockTime, *payload.BlockTime)
	assert.Equal(t, []string{"TxSig1"}, payload.TransactionSignatures)
}
