//go:build test

package store

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/store"
)

// These contracts have no behaviour of their own, so what is worth pinning down
// is that the generic embedding produces the method set each store should have.

type stubStore[D any] struct{}

func (stubStore[D]) Save(context.Context, D) error { return nil }

func (stubStore[D]) Supports(store.Type, reflect.Type) bool { return false }

type stubBlockStore struct {
	stubStore[event.BlockEvent]
}

func (stubBlockStore) GetLatest(context.Context, uuid.UUID) (uint64, error) {
	return 0, nil
}

type stubLatestBlockStore struct {
	stubStore[event.BlockEvent]
}

func (stubLatestBlockStore) Get(context.Context, uuid.UUID) (uint64, error) {
	return 0, nil
}

var (
	_ Store[event.TransactionEvent]                 = stubStore[event.TransactionEvent]{}
	_ EventStore[event.TransactionEvent]            = stubStore[event.TransactionEvent]{}
	_ TransactionEventStore[event.TransactionEvent] = stubStore[event.TransactionEvent]{}
	_ ContractEventStore[event.ContractEvent]       = stubStore[event.ContractEvent]{}
	_ BlockEventStore[event.BlockEvent]             = stubBlockStore{}
	_ LatestBlockStore[event.BlockEvent]            = stubLatestBlockStore{}
)

// Distinct event kinds must not satisfy each other's interface, which is what
// lets Supports route by data type.
func TestEventKindInterfacesAreDistinct(t *testing.T) {
	var transactionStore any = stubStore[event.TransactionEvent]{}

	_, isTransactionStore := transactionStore.(TransactionEventStore[event.TransactionEvent])
	assert.True(t, isTransactionStore)

	_, isContractStore := transactionStore.(ContractEventStore[event.ContractEvent])
	assert.False(t, isContractStore, "a transaction store must not pass as a contract store")

	_, isBlockStore := transactionStore.(BlockEventStore[event.SolanaBlockEvent])
	assert.False(t, isBlockStore, "a transaction store must not pass as a block store")

	_, isLatestBlockStore := transactionStore.(LatestBlockStore[event.SolanaBlockEvent])
	assert.False(t, isLatestBlockStore, "a transaction store must not pass as a latest block store")
}

// GetLatest is the only thing separating BlockEventStore from a plain
// EventStore, so it is worth pinning down that embedding still carries Save and
// Supports alongside it — and that nothing stored yet is reported through the
// boolean, not as an error.
func TestBlockEventStoreAddsGetLatest(t *testing.T) {
	var blockStore BlockEventStore[event.BlockEvent] = stubBlockStore{}

	latest, err := blockStore.GetLatest(context.Background(), uuid.New())
	assert.NoError(t, err)
	assert.Zero(t, latest)

	_, isPlainEventStore := any(blockStore).(EventStore[event.BlockEvent])
	assert.True(t, isPlainEventStore, "a BlockEventStore is still an EventStore")
}

func TestLatestBlockStoreAddsGet(t *testing.T) {
	var latestBlockStore LatestBlockStore[event.BlockEvent] = stubLatestBlockStore{}

	latest, err := latestBlockStore.Get(context.Background(), uuid.New())
	assert.NoError(t, err)
	assert.Zero(t, latest)

	_, isPlainEventStore := any(latestBlockStore).(EventStore[event.BlockEvent])
	assert.True(t, isPlainEventStore, "a LatestBlockStore is still an EventStore")
}
