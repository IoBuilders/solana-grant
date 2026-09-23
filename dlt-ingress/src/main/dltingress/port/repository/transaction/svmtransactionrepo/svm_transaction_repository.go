package svmtransactionrepo

import (
	"context"
	"dlt-ingress/src/main/dltingress/domain/transaction/svmtransaction"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/repository"
)

type Repository interface {
	repository.BaseRepository[Repository, svmtransaction.SvmTransaction]
	FindByTxId(ctx context.Context, txId string) (*svmtransaction.SvmTransaction, error)
}
