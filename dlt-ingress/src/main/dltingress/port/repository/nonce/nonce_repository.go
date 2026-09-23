package noncerepo

import (
	"context"
	"dlt-ingress/src/main/dltingress/domain/nonce"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/repository"
)

type Repository interface {
	repository.BaseRepository[Repository, nonce.Nonce]
	FindAndLockByDltAccountIdAndNetworkId(ctx context.Context, dltAccountId string, networkId string) (*nonce.Nonce, error)
	DeleteAll(ctx context.Context) error
}
