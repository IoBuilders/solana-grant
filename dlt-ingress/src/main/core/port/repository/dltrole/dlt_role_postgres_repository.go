package dltrolerepo

import (
	"context"
	"dlt-ingress/src/main/core/domain/role/dltrole"
	"fmt"

	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/db"
	"gorm.io/gorm"
)

type PostgresAssetDltRoleRepository struct {
	db *gorm.DB
	tm db.TransactionManager
}

func NewPostgresAssetDltRoleRepository(db *gorm.DB, tm db.TransactionManager) *PostgresAssetDltRoleRepository {
	return &PostgresAssetDltRoleRepository{
		db: db,
		tm: tm,
	}
}

func (r *PostgresAssetDltRoleRepository) FindByIds(ctx context.Context, ids []uuid.UUID) ([]dltrole.DltRole, error) {
	var entities []dltrole.DltRole
	result := r.getDB(ctx).Where("id IN (?)", ids).Find(&entities)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch dlt roles: %w", result.Error)
	}
	return entities, nil
}
