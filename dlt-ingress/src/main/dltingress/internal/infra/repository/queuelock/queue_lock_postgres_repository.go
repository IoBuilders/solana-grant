package queuelockrepo

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/domain/queuelock"
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

func (r *Postgres) FindAndLockByNetworkId(ctx context.Context, networkId string) (*queuelock.QueueLock, error) {
	var model QueueLock
	result := r.getDB(ctx).Where("network_id = ?", networkId).Clauses(clause.Locking{Strength: "UPDATE"}).First(&model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerrors.NewEntityNotFoundDomainError("QueueLock", networkId)
		}
		return nil, fmt.Errorf("failed to retrieve QueueLock: %w", result.Error)
	}
	entity := ToDomain(model)
	return &entity, nil
}

func (r *Postgres) CreateIfNotExists(ctx context.Context, entity *queuelock.QueueLock) error {

	result := r.getDB(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "network_id"}},
		DoNothing: true,
	}).Create(entity)
	return result.Error
}
