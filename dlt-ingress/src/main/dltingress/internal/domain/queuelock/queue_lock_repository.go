package queuelock

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/repository"
)

type Repository interface {
	repository.BaseRepository[Repository, QueueLock]
	FindAndLockByNetworkId(ctx context.Context, networkId string) (*QueueLock, error)
	CreateIfNotExists(ctx context.Context, entity *QueueLock) error
}
