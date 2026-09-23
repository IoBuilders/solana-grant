//go:build test

package eventmapper

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster/target"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

func newTestBroadcaster(t *testing.T, trg target.Target) broadcaster.Broadcaster {
	t.Helper()
	b, err := broadcaster.NewBroadcaster(uuid.New(), trg, uuid.New())
	require.NoError(t, err)
	return *b
}

func newBlockTarget(t *testing.T) target.Target {
	t.Helper()
	trg, err := target.NewBlockTarget([]target.Destination{"/webhook"})
	require.NoError(t, err)
	return trg
}

func TestBlockchainEventSourceToJsonMapper_Map_ContractEvent(t *testing.T) {
	filterID := uuid.New()
	filterTarget, err := target.NewFilterTarget([]target.Destination{"/webhook"}, filterID)
	require.NoError(t, err)
	b := newTestBroadcaster(t, filterTarget)

	e, err := event.NewSolanaContractEvent(uuid.New(), "Program1", "Sig1", 1, nil, "Transfer", event.ContractEventStatusConfirmed)
	require.NoError(t, err)

	m := NewEventToJsonMapper()
	result, err := m.Map(BlockchainEventSource{Event: e, Broadcaster: b})

	require.NoError(t, err)
	var payload ContractEventPayload
	require.NoError(t, json.Unmarshal(result, &payload))
	require.NotNil(t, payload.FilterID)
	assert.Equal(t, filterID, *payload.FilterID)
}

func TestBlockchainEventSourceToJsonMapper_Map_ContractEvent_NoFilterTarget(t *testing.T) {
	b := newTestBroadcaster(t, newBlockTarget(t))
	e, err := event.NewSolanaContractEvent(uuid.New(), "Program1", "Sig1", 1, nil, "Transfer", event.ContractEventStatusConfirmed)
	require.NoError(t, err)

	m := NewEventToJsonMapper()
	result, err := m.Map(BlockchainEventSource{Event: e, Broadcaster: b})

	require.NoError(t, err)
	var payload ContractEventPayload
	require.NoError(t, json.Unmarshal(result, &payload))
	assert.Nil(t, payload.FilterID)
}

func TestBlockchainEventSourceToJsonMapper_Map_BlockEvent(t *testing.T) {
	b := newTestBroadcaster(t, newBlockTarget(t))
	e, err := event.NewSolanaBlockEvent(uuid.New(), 1, "Blockhash1", nil, nil)
	require.NoError(t, err)

	m := NewEventToJsonMapper()
	result, err := m.Map(BlockchainEventSource{Event: e, Broadcaster: b})

	require.NoError(t, err)
	var payload BlockEventPayload
	require.NoError(t, json.Unmarshal(result, &payload))
	assert.Equal(t, uint64(1), payload.Slot)
	assert.Equal(t, "Blockhash1", payload.Blockhash)
}

func TestBlockchainEventSourceToJsonMapper_Map_TransactionEvent(t *testing.T) {
	b := newTestBroadcaster(t, newBlockTarget(t))
	e, err := event.NewSolanaTransactionEvent(uuid.New(), "TxSig1", 1, nil, nil, nil, nil)
	require.NoError(t, err)

	m := NewEventToJsonMapper()
	result, err := m.Map(BlockchainEventSource{Event: e, Broadcaster: b})

	require.NoError(t, err)
	var payload TransactionEventPayload
	require.NoError(t, json.Unmarshal(result, &payload))
	assert.Equal(t, "TxSig1", payload.Signature)
	assert.Equal(t, uint64(1), payload.Slot)
}

type unsupportedEvent struct{}

func (unsupportedEvent) EventType() event.Type { return event.TypeBlock }
func (unsupportedEvent) NodeID() uuid.UUID     { return uuid.New() }

func TestBlockchainEventSourceToJsonMapper_Map_UnsupportedEventType(t *testing.T) {
	b := newTestBroadcaster(t, newBlockTarget(t))

	m := NewEventToJsonMapper()
	_, err := m.Map(BlockchainEventSource{Event: unsupportedEvent{}, Broadcaster: b})

	assert.ErrorContains(t, err, "unsupported event type")
}
