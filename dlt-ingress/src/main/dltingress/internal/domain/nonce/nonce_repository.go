package nonce

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/repository"
)

type Repository interface {
	repository.BaseRepository[Repository, Nonce]
	FindAndLockByDltAccountIdAndNetworkId(ctx context.Context, dltAccountId string, networkId string) (*Nonce, error)
	DeleteAll(ctx context.Context) error
}
