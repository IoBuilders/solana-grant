package noncerepo

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/domain/nonce"
	"errors"
	"fmt"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/db"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Postgres struct {
	db *gorm.DB
	tm db.TransactionManager
}

func NewPostgres(gormDB *gorm.DB, tm db.TransactionManager) *Postgres {
	return &Postgres{db: gormDB, tm: tm}
}

func (r *Postgres) FindAndLockByDltAccountIdAndNetworkId(ctx context.Context, dltAccountId string, networkId string) (*nonce.Nonce, error) {
	var model Nonce
	result := r.getDB(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("dlt_account_id = ? AND network_id = ?", dltAccountId, networkId).
		First(&model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerrors.NewEntityNotFoundDomainError("Nonce", dltAccountId)
		}
		return nil, fmt.Errorf("failed to retrieve Nonce: %w", result.Error)
	}
	entity := ToDomain(model)
	return &entity, nil
}

func (r *Postgres) DeleteAll(ctx context.Context) error {
	result := r.getDB(ctx).Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&nonce.Nonce{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete all Nonce: %w", result.Error)
	}
	return nil
}
