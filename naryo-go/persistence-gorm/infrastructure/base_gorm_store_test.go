//go:build test

package infrastructure

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	appstore "gitlab.com/iobuilders/projects/eng/naryo-go/core/app/store"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/store"
)

func TestBase_Supports(t *testing.T) {
	s := NewBaseStore[event.TransactionEvent](nil)

	t.Run("MatchingBackendAndDataKind", func(t *testing.T) {
		assert.True(t, s.Supports(store.TypeGorm, reflect.TypeFor[event.TransactionEvent]()))
	})

	t.Run("WrongBackend", func(t *testing.T) {
		assert.False(t, s.Supports(store.Type("MONGO"), reflect.TypeFor[event.TransactionEvent]()))
		assert.False(t, s.Supports(store.Type(""), reflect.TypeFor[event.TransactionEvent]()))
	})

	t.Run("WrongDataKind", func(t *testing.T) {
		assert.False(t, s.Supports(store.TypeGorm, reflect.TypeFor[event.BlockEvent]()))
		assert.False(t, s.Supports(store.TypeGorm, reflect.TypeFor[event.ContractEvent]()))
	})

	t.Run("ConcreteChainTypeIsNotMatched", func(t *testing.T) {
		assert.False(t, s.Supports(store.TypeGorm, reflect.TypeFor[event.SolanaTransactionEvent]()))
	})

	t.Run("NilDataKind", func(t *testing.T) {
		assert.False(t, s.Supports(store.TypeGorm, nil))
	})
}

// Each instantiation answers for its own D, which is what lets a process holding
// several stores route by data kind.
func TestBase_SupportsIsPerInstantiation(t *testing.T) {
	transactions := NewBaseStore[event.TransactionEvent](nil)
	blocks := NewBaseStore[event.BlockEvent](nil)

	assert.True(t, transactions.Supports(store.TypeGorm, reflect.TypeFor[event.TransactionEvent]()))
	assert.False(t, blocks.Supports(store.TypeGorm, reflect.TypeFor[event.TransactionEvent]()))
	assert.True(t, blocks.Supports(store.TypeGorm, reflect.TypeFor[event.BlockEvent]()))
}

// --- fitness for purpose ---
//
// The helper only earns its place if embedding it, plus the operations a store
// implements itself, adds up to the core port. The assertions below make the
// compiler prove that: a drift between this Supports and Store's would break
// them, which no runtime test would catch.

// stubTransactionStore is the minimum a concrete store adds on top of the helper.
type stubTransactionStore struct {
	BaseStore[event.TransactionEvent]
}

func (stubTransactionStore) Save(context.Context, event.TransactionEvent) error {
	return nil
}

// stubBlockStore covers the one store shape that adds an operation of its own.
type stubBlockStore struct {
	BaseStore[event.BlockEvent]
}

func (stubBlockStore) Save(context.Context, event.BlockEvent) error {
	return nil
}

func (stubBlockStore) GetLatest(context.Context, uuid.UUID) (uint64, error) {
	return 0, nil
}

var (
	_ appstore.TransactionEventStore[event.TransactionEvent] = stubTransactionStore{}
	_ appstore.BlockEventStore[event.BlockEvent]             = stubBlockStore{}
)

// Supports must still route correctly when reached through the port rather than
// through the concrete type.
func TestBase_SupportsThroughPort(t *testing.T) {
	db := &gorm.DB{}
	var s appstore.TransactionEventStore[event.TransactionEvent] = stubTransactionStore{
		BaseStore: NewBaseStore[event.TransactionEvent](db),
	}

	assert.True(t, s.Supports(store.TypeGorm, reflect.TypeFor[event.TransactionEvent]()))
	assert.False(t, s.Supports(store.TypeGorm, reflect.TypeFor[event.BlockEvent]()))
	assert.False(t, s.Supports(store.Type("MONGO"), reflect.TypeFor[event.TransactionEvent]()))
}

// The embedded db reaches the concrete store, so it does not need a field of its
// own.
func TestBase_ProvidesDBToEmbedder(t *testing.T) {
	db := &gorm.DB{}
	s := stubTransactionStore{BaseStore: NewBaseStore[event.TransactionEvent](db)}

	assert.Same(t, db, s.DB())
}
