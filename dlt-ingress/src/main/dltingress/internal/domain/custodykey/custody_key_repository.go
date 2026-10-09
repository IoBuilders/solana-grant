package custodykey

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/repository"
)

type Repository interface {
	repository.BaseRepository[Repository, CustodyKey]
	FindByDltAccountId(ctx context.Context, dltAccountId string) (*CustodyKey, error)
	FindByDltAccountIds(ctx context.Context, dltAccountIds []string) ([]*CustodyKey, error)
}
