package evmtransaction

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/repository"
)

type Repository interface {
	repository.BaseRepository[Repository, EvmTransaction]
	FindByTxId(ctx context.Context, txId string) (*EvmTransaction, error)
	HardDelete(ctx context.Context, tx *EvmTransaction) error
}
