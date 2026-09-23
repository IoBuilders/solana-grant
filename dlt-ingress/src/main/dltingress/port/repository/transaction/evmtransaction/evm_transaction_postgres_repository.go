package evmtransactionrepo

import (
	"context"
	"dlt-ingress/src/main/dltingress/domain/transaction/evmtransaction"
	"errors"
	"fmt"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/db"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"

	"gorm.io/gorm"
)

type PostgresEvmTransactionRepository struct {
	db *gorm.DB
	tm db.TransactionManager
}

func NewPostgresEvmTransactionRepository(gormDB *gorm.DB, tm db.TransactionManager) *PostgresEvmTransactionRepository {
	return &PostgresEvmTransactionRepository{db: gormDB, tm: tm}
}

func (r *PostgresEvmTransactionRepository) FindByTxId(ctx context.Context, txId string) (*evmtransaction.EvmTransaction, error) {
	var entity evmtransaction.EvmTransaction
	result := r.getDB(ctx).Where("tx_id = ?", txId).First(&entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerrors.NewEntityNotFoundDomainError("EvmTransaction", txId)
		}
		return nil, fmt.Errorf("failed to retrieve Evm transaction: %w", result.Error)
	}
	return &entity, nil
}
