//go:build test

package subscription

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/interactor"
	appstore "gitlab.com/iobuilders/projects/eng/naryo-go/core/app/store"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/block"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/node"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/store"
)

// --- stubs ---

// fakeBlockInteractor is a hand-written test double for interactor.BlockInteractor.
type fakeBlockInteractor struct {
	slot    uint64
	slotErr error
}

func (i *fakeBlockInteractor) Type() interactor.Type { return interactor.TypeBlock }

func (i *fakeBlockInteractor) GetBlock(context.Context, uint64) (*block.SolanaBlock, error) {
	return nil, nil
}

func (i *fakeBlockInteractor) GetSlot(context.Context) (uint64, error) {
	if i.slotErr != nil {
		return 0, i.slotErr
	}
	return i.slot, nil
}

// fakeLatestBlockStore is a hand-written test double for
// store.LatestBlockStore[event.SolanaBlockEvent]. GetStartSlot only ever
// calls Get, so Save/Supports are unused no-ops.
type fakeLatestBlockStore struct {
	latestSlot uint64
	latestErr  error
}

func (s *fakeLatestBlockStore) Save(context.Context, event.SolanaBlockEvent) error { return nil }

func (s *fakeLatestBlockStore) Supports(store.Type, reflect.Type) bool { return true }

func (s *fakeLatestBlockStore) Get(context.Context, uuid.UUID) (uint64, error) {
	if s.latestErr != nil {
		return 0, s.latestErr
	}
	return s.latestSlot, nil
}

var _ appstore.LatestBlockStore[event.SolanaBlockEvent] = (*fakeLatestBlockStore)(nil)

// --- tests ---

func TestSolanaStartSlotCalculator_GetStartSlot_NoStoredSlot_UsesInteractorSlot(t *testing.T) {
	store := &fakeLatestBlockStore{}

	n := &node.Node{ID: uuid.New()}
	c := NewSolanaStartSlotCalculator(n, &fakeBlockInteractor{slot: 123}, store)

	slot, err := c.GetStartSlot(context.Background())

	require.NoError(t, err)
	assert.Equal(t, uint64(123), slot)
}

func TestSolanaStartSlotCalculator_GetStartSlot_StoredSlot_ReturnsNextSlot(t *testing.T) {
	store := &fakeLatestBlockStore{latestSlot: 99}

	n := &node.Node{ID: uuid.New()}
	c := NewSolanaStartSlotCalculator(n, &fakeBlockInteractor{}, store)

	slot, err := c.GetStartSlot(context.Background())

	require.NoError(t, err)
	assert.Equal(t, uint64(100), slot)
}

func TestSolanaStartSlotCalculator_GetStartSlot_StoreError_ReturnsError(t *testing.T) {
	dbErr := errors.New("connection reset")
	store := &fakeLatestBlockStore{latestErr: dbErr}

	n := &node.Node{ID: uuid.New()}
	c := NewSolanaStartSlotCalculator(n, &fakeBlockInteractor{}, store)

	slot, err := c.GetStartSlot(context.Background())

	assert.ErrorIs(t, err, dbErr)
	assert.Equal(t, uint64(0), slot)
}

func TestSolanaStartSlotCalculator_GetStartSlot_InteractorError_ReturnsError(t *testing.T) {
	store := &fakeLatestBlockStore{}
	interactorErr := errors.New("rpc failed")

	n := &node.Node{ID: uuid.New()}
	c := NewSolanaStartSlotCalculator(n, &fakeBlockInteractor{slotErr: interactorErr}, store)

	slot, err := c.GetStartSlot(context.Background())

	assert.ErrorIs(t, err, interactorErr)
	assert.Equal(t, uint64(0), slot)
}
