package store

import (
	"context"
	"reflect"

	"github.com/google/uuid"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/store"
)

// NoopEventStore satisfies Store[D] by persisting nothing. It backs a Node
// with no configured store for a given event kind, so a trigger can depend
// on the port unconditionally instead of checking whether one is configured.
type NoopEventStore[D any] struct{}

func (NoopEventStore[D]) Save(context.Context, D) error { return nil }

func (NoopEventStore[D]) Supports(store.Type, reflect.Type) bool { return false }

// NoopBlockEventStore is the BlockEventStore counterpart of NoopEventStore,
// adding a GetLatest that always reports nothing stored.
type NoopBlockEventStore struct {
	NoopEventStore[event.SolanaBlockEvent]
}

func (NoopBlockEventStore) GetLatest(context.Context, uuid.UUID) (uint64, error) {
	return 0, nil
}

var (
	_ BlockEventStore[event.SolanaBlockEvent]             = NoopBlockEventStore{}
	_ TransactionEventStore[event.SolanaTransactionEvent] = NoopEventStore[event.SolanaTransactionEvent]{}
	_ ContractEventStore[event.SolanaContractEvent]       = NoopEventStore[event.SolanaContractEvent]{}
)
