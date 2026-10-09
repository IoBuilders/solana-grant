package failedtransactionrepo

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction/failedtransaction"
	"errors"
	"fmt"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api/pagination"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/db"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"
	"gorm.io/gorm"
)

type Postgres struct {
	db *gorm.DB
	tm db.TransactionManager
}

func NewPostgresFailedTransactionRepository(gormDB *gorm.DB, tm db.TransactionManager) *Postgres {
	return &Postgres{db: gormDB, tm: tm}
}

func (r *Postgres) FindByTxId(ctx context.Context, txId string) (*failedtransaction.FailedTransaction, error) {
	var model FailedTransaction
	result := r.getDB(ctx).Where("tx_id = ?", txId).First(&model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerrors.NewEntityNotFoundDomainError("EvmTransaction", txId)
		}
		return nil, fmt.Errorf("failed to retrieve failed transaction: %w", result.Error)
	}
	entity := ToDomain(model)
	return &entity, nil
}

func (p *Postgres) FindByStatusPaginated(ctx context.Context, status failedtransaction.Status, params pagination.PaginationParams) ([]failedtransaction.FailedTransaction, error) {
	var models []FailedTransaction
	result := p.getDB(ctx).Model(&FailedTransaction{}).
		Where("status = ?", status).
		Limit(params.PageSize).
		Offset(params.Offset).
		Order(params.OrderClause()).
		Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch FailedTransaction entities: %w", result.Error)
	}
	return ToDomainList(models), nil
}

func (p *Postgres) CountAllByFilters(ctx context.Context, status failedtransaction.Status) (int, error) {
	var count int64
	result := p.getDB(ctx).Model(&FailedTransaction{}).
		Where("status = ?", status).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count FailedTransaction entities: %w", result.Error)
	}
	return int(count), nil
}

func (r *Postgres) ExistByTxIdAndStatus(ctx context.Context, txId string, status failedtransaction.Status) (bool, error) {
	var res int
	err := r.getDB(ctx).
		Model(&FailedTransaction{}).
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
