package eventstorerepo

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/domain/eventstore"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/repository"
)

type Repository interface {
	repository.BaseRepository[Repository, eventstore.EventStore]
	ClaimPendingEvents(ctx context.Context, isCross bool, limit int) ([]eventstore.EventStore, error)
	FindTraceParentByTxHash(ctx context.Context, txHash string) (string, error)
}
