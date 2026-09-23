package evmtransactionrepo

import (
	"context"
	"dlt-ingress/src/main/dltingress/domain/transaction/evmtransaction"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/repository"
)

type Repository interface {
	repository.BaseRepository[Repository, evmtransaction.EvmTransaction]
	FindByTxId(ctx context.Context, txId string) (*evmtransaction.EvmTransaction, error)
}
