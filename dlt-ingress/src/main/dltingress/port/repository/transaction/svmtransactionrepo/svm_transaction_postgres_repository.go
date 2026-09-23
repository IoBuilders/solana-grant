package svmtransactionrepo

import (
	"context"
	"dlt-ingress/src/main/dltingress/domain/transaction/svmtransaction"
	"errors"
	"fmt"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/db"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"

	"gorm.io/gorm"
)

type PostgresSVMTransactionRepository struct {
	db *gorm.DB
	tm db.TransactionManager
}

func NewPostgresSVMTransactionRepository(gormDB *gorm.DB, tm db.TransactionManager) *PostgresSVMTransactionRepository {
	return &PostgresSVMTransactionRepository{db: gormDB, tm: tm}
}

func (r *PostgresSVMTransactionRepository) FindByTxId(ctx context.Context, txId string) (*svmtransaction.SvmTransaction, error) {
	var entity svmtransaction.SvmTransaction
	result := r.getDB(ctx).Where("tx_id = ?", txId).First(&entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerrors.NewEntityNotFoundDomainError("SvmTransaction", txId)
		}
		return nil, fmt.Errorf("failed to retrieve SVM transaction: %w", result.Error)
	}
	return &entity, nil
}
