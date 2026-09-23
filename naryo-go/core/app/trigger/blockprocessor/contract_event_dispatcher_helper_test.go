//go:build test

package blockprocessor

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
)

func TestContractEventDispatcherHelper_Execute_DispatchesContractEvent(t *testing.T) {
	nodeID := uuid.New()
	dispatcher := &fakeDispatcher{}
	helper := NewContractEventDispatcherHelper(dispatcher)

	err := helper.Execute(
		context.Background(),
		nodeID,
		"program-id",
		"signature",
		42,
		[]parameter.ContractEventParameter{},
		"EventName",
		event.ContractEventStatusConfirmed,
	)

	require.NoError(t, err)
	require.Len(t, dispatcher.dispatched, 1)

	contractEvent, ok := dispatcher.dispatched[0].(event.SolanaContractEvent)
	require.True(t, ok)
	assert.Equal(t, nodeID, contractEvent.NodeID())
	assert.Equal(t, "program-id", contractEvent.ProgramID)
	assert.Equal(t, "signature", contractEvent.Signature)
	assert.Equal(t, uint64(42), contractEvent.Slot)
	assert.Equal(t, "EventName", contractEvent.EventName())
	assert.Equal(t, event.ContractEventStatusConfirmed, contractEvent.Status())
}

func TestContractEventDispatcherHelper_Execute_InvalidData_ReturnsErrorWithoutDispatching(t *testing.T) {
	dispatcher := &fakeDispatcher{}
	helper := NewContractEventDispatcherHelper(dispatcher)

	err := helper.Execute(
		context.Background(),
		uuid.New(),
		"",
		"signature",
		42,
		nil,
		"EventName",
		event.ContractEventStatusConfirmed,
	)

	require.Error(t, err)
	assert.Empty(t, dispatcher.dispatched)
}
