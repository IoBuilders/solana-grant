package txqueueslotrepo

import (
	"context"
	"dlt-ingress/src/main/dltingress/domain/txqueueslot"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/repository"
)

type Repository interface {
	repository.BaseRepository[Repository, txqueueslot.TxQueueSlot]
	CountByNetworkIdAndNotExpired(ctx context.Context, networkId string) (int, error)
	DeleteExpired(ctx context.Context) error
	DeleteFirstByNetworkId(ctx context.Context, networkId string) error
}
