package failedtransactionrepo

import (
	"context"
	"dlt-ingress/src/main/dltingress/domain/transaction/failedtransaction"
	"errors"
	"fmt"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api/pagination"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/db"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"
	"gorm.io/gorm"
)

type PostgresFailedTransactionRepository struct {
	db *gorm.DB
	tm db.TransactionManager
}

func NewPostgresFailedTransactionRepository(gormDB *gorm.DB, tm db.TransactionManager) *PostgresFailedTransactionRepository {
	return &PostgresFailedTransactionRepository{db: gormDB, tm: tm}
}

func (r *PostgresFailedTransactionRepository) FindByTxId(ctx context.Context, txId string) (*failedtransaction.FailedTransaction, error) {
	var entity failedtransaction.FailedTransaction
	result := r.getDB(ctx).Where("tx_id = ?", txId).First(&entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerrors.NewEntityNotFoundDomainError("EvmTransaction", txId)
		}
		return nil, fmt.Errorf("failed to retrieve failed transaction: %w", result.Error)
	}
	return &entity, nil
}

func (p *PostgresFailedTransactionRepository) FindByStatusPaginated(ctx context.Context, status failedtransaction.Status, params pagination.PaginationParams) ([]failedtransaction.FailedTransaction, error) {
	var entities []failedtransaction.FailedTransaction
	result := p.getDB(ctx).Model(&failedtransaction.FailedTransaction{}).
		Where("status = ?", status).
		Limit(params.PageSize).
		Offset(params.Offset).
		Order(params.OrderClause()).
		Find(&entities)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch FailedTransaction entities: %w", result.Error)
	}
	return entities, nil
}

func (p *PostgresFailedTransactionRepository) CountAllByFilters(ctx context.Context, status failedtransaction.Status) (int, error) {
	var count int64
	result := p.getDB(ctx).Model(&failedtransaction.FailedTransaction{}).
		Where("status = ?", status).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count FailedTransaction entities: %w", result.Error)
	}
	return int(count), nil
}

func (r *PostgresFailedTransactionRepository) ExistByTxIdAndStatus(ctx context.Context, txId string, status failedtransaction.Status) (bool, error) {
	var res int
	err := r.getDB(ctx).
		Model(&failedtransaction.FailedTransaction{}).
		Select("1").
		Where("tx_id = ? AND status = ?", txId, status).
		Take(&res).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check existence of FailedTransaction: %w", err)
	}
	return true, nil
}
