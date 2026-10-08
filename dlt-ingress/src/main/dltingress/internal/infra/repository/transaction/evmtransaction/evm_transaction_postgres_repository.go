package evmtransactionrepo

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction/evmtransaction"
	"errors"
	"fmt"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/db"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"

	"gorm.io/gorm"
)

type Postgres struct {
	db *gorm.DB
	tm db.TransactionManager
}

func NewPostgres(gormDB *gorm.DB, tm db.TransactionManager) *Postgres {
	return &Postgres{db: gormDB, tm: tm}
}

func (r *Postgres) FindByTxId(ctx context.Context, txId string) (*evmtransaction.EvmTransaction, error) {
	var model EvmTransaction
	result := r.getDB(ctx).Where("tx_id = ?", txId).First(&model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerrors.NewEntityNotFoundDomainError("EvmTransaction", txId)
		}
		return nil, fmt.Errorf("failed to retrieve Evm transaction: %w", result.Error)
	}
	entity := ToDomain(model)
	return &entity, nil
}

func (r *Postgres) HardDelete(ctx context.Context, evm *evmtransaction.EvmTransaction) error {
	rec := FromDomain(*evm)
	if err := r.getDB(ctx).Unscoped().Delete(&rec).Error; err != nil {
		return fmt.Errorf("failed to hard delete EvmTransaction: %w", err)
	}
	return nil
}
