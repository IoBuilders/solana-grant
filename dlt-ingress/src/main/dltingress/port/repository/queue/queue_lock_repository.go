package queuelockrepo

import (
	"context"
	"dlt-ingress/src/main/dltingress/domain/queuelock"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/repository"
)

type Repository interface {
	repository.BaseRepository[Repository, queuelock.QueueLock]
	FindAndLockByNetworkId(ctx context.Context, networkId string) (*queuelock.QueueLock, error)
	CreateIfNotExists(ctx context.Context, entity *queuelock.QueueLock) error
}
