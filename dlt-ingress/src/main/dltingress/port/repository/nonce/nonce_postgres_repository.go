package noncerepo

import (
	"context"
	"dlt-ingress/src/main/dltingress/domain/nonce"
	"errors"
	"fmt"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/db"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PostgresNonceRepository struct {
	db *gorm.DB
	tm db.TransactionManager
}

func NewPostgresNonceRepository(gormDB *gorm.DB, tm db.TransactionManager) *PostgresNonceRepository {
	return &PostgresNonceRepository{db: gormDB, tm: tm}
}

func (r *PostgresNonceRepository) FindAndLockByDltAccountIdAndNetworkId(ctx context.Context, dltAccountId string, networkId string) (*nonce.Nonce, error) {
	var entity nonce.Nonce
	result := r.getDB(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("dlt_account_id = ? AND network_id = ?", dltAccountId, networkId).
		First(&entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerrors.NewEntityNotFoundDomainError("Nonce", dltAccountId)
		}
		return nil, fmt.Errorf("failed to retrieve Nonce: %w", result.Error)
	}
	return &entity, nil
}

func (r *PostgresNonceRepository) DeleteAll(ctx context.Context) error {
	result := r.getDB(ctx).Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&nonce.Nonce{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete all Nonce: %w", result.Error)
	}
	return nil
}
