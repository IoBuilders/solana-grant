package dltrolerepo

import (
	"context"
	"dlt-ingress/src/main/core/domain/role/dltrole"

	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/repository"
)

type Repository interface {
	repository.BaseRepository[Repository, dltrole.DltRole]
	FindByIds(ctx context.Context, ids []uuid.UUID) ([]dltrole.DltRole, error)
}
