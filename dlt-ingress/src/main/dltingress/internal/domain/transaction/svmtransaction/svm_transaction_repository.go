package svmtransaction

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/repository"
)

type Repository interface {
	repository.BaseRepository[Repository, SvmTransaction]
	FindByTxId(ctx context.Context, txId string) (*SvmTransaction, error)
	ExistsByOriginalTxId(ctx context.Context, originalTxId string) (bool, error)
	HardDelete(ctx context.Context, tx *SvmTransaction) error
}
