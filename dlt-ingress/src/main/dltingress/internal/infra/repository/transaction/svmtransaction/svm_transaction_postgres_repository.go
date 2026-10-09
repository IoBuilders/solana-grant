package svmtransactionrepo

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction/svmtransaction"
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

func (r *Postgres) FindByTxId(ctx context.Context, txId string) (*svmtransaction.SvmTransaction, error) {
	var model SvmTransaction
	result := r.getDB(ctx).Where("tx_id = ?", txId).First(&model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerrors.NewEntityNotFoundDomainError("SvmTransaction", txId)
		}
		return nil, fmt.Errorf("failed to retrieve SVM transaction: %w", result.Error)
	}
	entity := ToDomain(model)
	return &entity, nil
}

func (r *Postgres) HardDelete(ctx context.Context, svm *svmtransaction.SvmTransaction) error {
	rec := FromDomain(*svm)
	if err := r.getDB(ctx).Unscoped().Delete(&rec).Error; err != nil {
		return fmt.Errorf("failed to hard delete SvmTransaction: %w", err)
	}
	return nil
}
