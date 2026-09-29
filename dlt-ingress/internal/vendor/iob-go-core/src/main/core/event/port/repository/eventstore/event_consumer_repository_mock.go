package eventstorerepo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api/pagination"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/domain/eventstore"
)

type EventConsumerRepositoryMock struct {
	mock.Mock
}

func (m *EventConsumerRepositoryMock) Create(ctx context.Context, entity *eventstore.EventConsumer) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *EventConsumerRepositoryMock) Save(ctx context.Context, entity *eventstore.EventConsumer) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *EventConsumerRepositoryMock) Delete(ctx context.Context, entity *eventstore.EventConsumer) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *EventConsumerRepositoryMock) FindById(ctx context.Context, id uuid.UUID) (*eventstore.EventConsumer, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventstore.EventConsumer), args.Error(1)
}

func (m *EventConsumerRepositoryMock) FindByIdPreload(ctx context.Context, id uuid.UUID) (*eventstore.EventConsumer, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*eventstore.EventConsumer), args.Error(1)
}

func (m *EventConsumerRepositoryMock) FindAll(ctx context.Context) ([]eventstore.EventConsumer, error) {
	args := m.Called(ctx)
	return args.Get(0).([]eventstore.EventConsumer), args.Error(1)
}

func (m *EventConsumerRepositoryMock) FindAllWithDeleted(ctx context.Context) ([]eventstore.EventConsumer, error) {
	args := m.Called(ctx)
	return args.Get(0).([]eventstore.EventConsumer), args.Error(1)
}

func (m *EventConsumerRepositoryMock) FindAllPreload(ctx context.Context) ([]eventstore.EventConsumer, error) {
	args := m.Called(ctx)
	return args.Get(0).([]eventstore.EventConsumer), args.Error(1)
}

func (m *EventConsumerRepositoryMock) FindAllPaginated(ctx context.Context, params pagination.PaginationParams) ([]eventstore.EventConsumer, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]eventstore.EventConsumer), args.Error(1)
}

func (m *EventConsumerRepositoryMock) FindAllPaginatedPreload(ctx context.Context, params pagination.PaginationParams) ([]eventstore.EventConsumer, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]eventstore.EventConsumer), args.Error(1)
}

func (m *EventConsumerRepositoryMock) CountAll(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Get(0).(int), args.Error(1)
}

func (m *EventConsumerRepositoryMock) UpdateStuckEventConsumers(ctx context.Context, cutoff time.Time) (int64, error) {
	args := m.Called(ctx, cutoff)
	return args.Get(0).(int64), args.Error(1)
}

func (m *EventConsumerRepositoryMock) ExistById(ctx context.Context, id uuid.UUID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *EventConsumerRepositoryMock) FindByStatusPaginated(ctx context.Context, status eventstore.Status, params pagination.PaginationParams) ([]eventstore.EventConsumer, error) {
	args := m.Called(ctx, status, params)
	return args.Get(0).([]eventstore.EventConsumer), args.Error(1)
}

func (m *EventConsumerRepositoryMock) CountAllByFilters(ctx context.Context, status eventstore.Status) (int, error) {
	args := m.Called(ctx, status)
	return args.Int(0), args.Error(1)
}

var _ EventConsumerRepository = (*EventConsumerRepositoryMock)(nil)
