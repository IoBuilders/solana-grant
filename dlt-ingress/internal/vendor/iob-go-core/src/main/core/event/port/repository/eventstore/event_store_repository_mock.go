package eventstorerepo

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api/pagination"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/domain/eventstore"
)

type EventStoreRepositoryMock struct {
	mock.Mock
}

func (m *EventStoreRepositoryMock) Create(ctx context.Context, entity *eventstore.EventStore) error {
	args := m.Called(ctx, entity)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	return nil
}

func (m *EventStoreRepositoryMock) Save(ctx context.Context, entity *eventstore.EventStore) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *EventStoreRepositoryMock) Delete(ctx context.Context, entity *eventstore.EventStore) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *EventStoreRepositoryMock) FindById(ctx context.Context, id uuid.UUID) (*eventstore.EventStore, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventstore.EventStore), args.Error(1)
}

func (m *EventStoreRepositoryMock) FindByIdPreload(ctx context.Context, id uuid.UUID) (*eventstore.EventStore, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventstore.EventStore), args.Error(1)
}

func (m *EventStoreRepositoryMock) FindAll(ctx context.Context) ([]eventstore.EventStore, error) {
	args := m.Called(ctx)
	return args.Get(0).([]eventstore.EventStore), args.Error(1)
}

func (m *EventStoreRepositoryMock) FindAllWithDeleted(ctx context.Context) ([]eventstore.EventStore, error) {
	args := m.Called(ctx)
	return args.Get(0).([]eventstore.EventStore), args.Error(1)
}

func (m *EventStoreRepositoryMock) FindAllPreload(ctx context.Context) ([]eventstore.EventStore, error) {
	args := m.Called(ctx)
	return args.Get(0).([]eventstore.EventStore), args.Error(1)
}

func (m *EventStoreRepositoryMock) FindAllPaginated(ctx context.Context, params pagination.PaginationParams) ([]eventstore.EventStore, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]eventstore.EventStore), args.Error(1)
}

func (m *EventStoreRepositoryMock) FindAllPaginatedPreload(ctx context.Context, params pagination.PaginationParams) ([]eventstore.EventStore, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]eventstore.EventStore), args.Error(1)
}

func (m *EventStoreRepositoryMock) CountAll(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Get(0).(int), args.Error(1)
}

func (m *EventStoreRepositoryMock) ClaimPendingEvents(ctx context.Context, isCross bool, limit int) ([]eventstore.EventStore, error) {
	args := m.Called(ctx, isCross, limit)
	return args.Get(0).([]eventstore.EventStore), args.Error(1)
}

func (m *EventStoreRepositoryMock) FindTraceParentByTxHash(ctx context.Context, txHash string) (string, error) {
	args := m.Called(ctx, txHash)
	return args.String(0), args.Error(1)
}

func (m *EventStoreRepositoryMock) ExistById(ctx context.Context, id uuid.UUID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

var _ Repository = (*EventStoreRepositoryMock)(nil)
