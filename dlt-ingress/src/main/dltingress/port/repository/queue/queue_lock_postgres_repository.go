package queuelockrepo

import (
	"context"
	"dlt-ingress/src/main/dltingress/domain/queuelock"
	"errors"
	"fmt"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/db"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PostgresQueueLockRepository struct {
	db *gorm.DB
	tm db.TransactionManager
}

func NewPostgresQueueLockRepository(gormDB *gorm.DB, tm db.TransactionManager) *PostgresQueueLockRepository {
	return &PostgresQueueLockRepository{db: gormDB, tm: tm}
}

func (r *PostgresQueueLockRepository) FindAndLockByNetworkId(ctx context.Context, networkId string) (*queuelock.QueueLock, error) {
	var entity queuelock.QueueLock
	result := r.getDB(ctx).Where("network_id = ?", networkId).Clauses(clause.Locking{Strength: "UPDATE"}).First(&entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerrors.NewEntityNotFoundDomainError("QueueLock", networkId)
		}
		return nil, fmt.Errorf("failed to retrieve QueueLock: %w", result.Error)
	}
	return &entity, nil
}

func (r *PostgresQueueLockRepository) CreateIfNotExists(ctx context.Context, entity *queuelock.QueueLock) error {
	result := r.getDB(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "network_id"}},
		DoNothing: true,
	}).Create(entity)
	return result.Error
}
