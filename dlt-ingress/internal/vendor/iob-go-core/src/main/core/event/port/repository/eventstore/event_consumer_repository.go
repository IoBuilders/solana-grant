package eventstorerepo

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api/pagination"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/domain/eventstore"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/repository"

	"time"
)

type EventConsumerRepository interface {
	repository.BaseRepository[EventConsumerRepository, eventstore.EventConsumer]
	UpdateStuckEventConsumers(ctx context.Context, cutoff time.Time) (int64, error)
	FindByStatusPaginated(ctx context.Context, status eventstore.Status, params pagination.PaginationParams) ([]eventstore.EventConsumer, error)
	CountAllByFilters(ctx context.Context, status eventstore.Status) (int, error)
}
