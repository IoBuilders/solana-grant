package custodykeyrepo

import (
	"context"
	"errors"
	"fmt"

	"dlt-ingress/src/main/dltingress/domain/custodykey"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/db"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"
	"gorm.io/gorm"
)

type PostgresCustodyKeyRepository struct {
	db *gorm.DB
	tm db.TransactionManager
}

func NewPostgresCustodyKeyRepository(gormDB *gorm.DB, tm db.TransactionManager) *PostgresCustodyKeyRepository {
	return &PostgresCustodyKeyRepository{db: gormDB, tm: tm}
}

func (r *PostgresCustodyKeyRepository) FindByDltAccountId(ctx context.Context, dltAccountId string) (*custodykey.CustodyKey, error) {
	var entity custodykey.CustodyKey
	result := r.getDB(ctx).Where("dlt_account_id = ?", dltAccountId).First(&entity)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerrors.NewEntityNotFoundDomainError("CustodyKey", dltAccountId)
		}
		return nil, fmt.Errorf("failed to retrieve CustodyKey: %w", result.Error)
	}
	return &entity, nil
}

func (r *PostgresCustodyKeyRepository) FindByDltAccountIds(ctx context.Context, dltAccountIds []string) ([]*custodykey.CustodyKey, error) {
	var entities []*custodykey.CustodyKey
	if err := r.getDB(ctx).Where("dlt_account_id IN ?", dltAccountIds).Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve CustodyKeys: %w", err)
	}
	return entities, nil
}
