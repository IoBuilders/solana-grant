package faucetwallet

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/repository"
)

type Repository interface {
	repository.BaseRepository[Repository, FaucetWallet]
	FindByNetworkId(ctx context.Context, networkId string) (*FaucetWallet, error)
	ExistByNetworkId(ctx context.Context, networkId string) (bool, error)
}
