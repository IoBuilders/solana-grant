package custodykeyrepo

import (
	"context"
	"dlt-ingress/src/main/dltingress/domain/custodykey"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/repository"
)

type Repository interface {
	repository.BaseRepository[Repository, custodykey.CustodyKey]
	FindByDltAccountId(ctx context.Context, dltAccountId string) (*custodykey.CustodyKey, error)
	FindByDltAccountIds(ctx context.Context, dltAccountIds []string) ([]*custodykey.CustodyKey, error)
}
