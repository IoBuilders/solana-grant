package custodykeyrepo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"dlt-ingress/src/main/dltingress/internal/domain/custodykey"

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

func (r *Postgres) FindByDltAccountId(ctx context.Context, dltAccountId string) (*custodykey.CustodyKey, error) {
	//TODO when using LOWER index is not used, making the query less efficient. Normalize account in creation or apply another solution
	var model CustodyKey
	result := r.getDB(ctx).Where("LOWER(dlt_account_id) = LOWER(?)", dltAccountId).First(&model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerrors.NewEntityNotFoundDomainError("CustodyKey", dltAccountId)
		}
		return nil, fmt.Errorf("failed to retrieve CustodyKey: %w", result.Error)
	}
	entity := ToDomain(model)
	return &entity, nil
}

func (r *Postgres) FindByDltAccountIds(ctx context.Context, dltAccountIds []string) ([]*custodykey.CustodyKey, error) {
	//TODO when using LOWER index is not used, making the query less efficient. Normalize account in creation or apply another solution
	var models []*CustodyKey
	lowerDltAccountIds := make([]string, len(dltAccountIds))
	for i, dltAccountId := range dltAccountIds {
		lowerDltAccountIds[i] = strings.ToLower(dltAccountId)
	}
	if err := r.getDB(ctx).Where("LOWER(dlt_account_id) IN ?", lowerDltAccountIds).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve CustodyKeys: %w", err)
	}
	return ToDomainPointerList(models), nil
}
