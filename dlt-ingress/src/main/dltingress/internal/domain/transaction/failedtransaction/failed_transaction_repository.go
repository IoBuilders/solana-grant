package failedtransaction

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api/pagination"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/repository"
)

type Repository interface {
	repository.BaseRepository[Repository, FailedTransaction]
	FindByTxId(ctx context.Context, txId string) (*FailedTransaction, error)
	FindByStatusPaginated(ctx context.Context, status Status, params pagination.PaginationParams) ([]FailedTransaction, error)
	CountAllByFilters(ctx context.Context, status Status) (int, error)
	ExistByTxIdAndStatus(ctx context.Context, txId string, status Status) (bool, error)
}
