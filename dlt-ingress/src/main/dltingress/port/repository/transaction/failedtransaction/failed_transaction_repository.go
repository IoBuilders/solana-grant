package failedtransactionrepo

import (
	"context"
	"dlt-ingress/src/main/dltingress/domain/transaction/failedtransaction"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api/pagination"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/repository"
)

type Repository interface {
	repository.BaseRepository[Repository, failedtransaction.FailedTransaction]
	FindByTxId(ctx context.Context, txId string) (*failedtransaction.FailedTransaction, error)
	FindByStatusPaginated(ctx context.Context, status failedtransaction.Status, params pagination.PaginationParams) ([]failedtransaction.FailedTransaction, error)
	CountAllByFilters(ctx context.Context, status failedtransaction.Status) (int, error)
	ExistByTxIdAndStatus(ctx context.Context, txId string, status failedtransaction.Status) (bool, error)
}
