package faucetwalletrepo

import (
	"context"
	"errors"
	"fmt"

	"dlt-ingress/src/main/dltingress/internal/domain/domainerrors"
	"dlt-ingress/src/main/dltingress/internal/domain/faucetwallet"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/db"
	"gorm.io/gorm"
)

type Postgres struct {
	db *gorm.DB
	tm db.TransactionManager
}

func NewPostgres(gormDB *gorm.DB, tm db.TransactionManager) *Postgres {
	return &Postgres{db: gormDB, tm: tm}
}

func (r *Postgres) FindByNetworkId(ctx context.Context, networkId string) (*faucetwallet.FaucetWallet, error) {
	var model FaucetWallet
	result := r.getDB(ctx).Where("network_id = ?", networkId).First(&model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domainerrors.NewFaucetWalletNotFoundByNetworkIdDomainError(networkId)
		}
		return nil, fmt.Errorf("failed to retrieve FaucetWallet: %w", result.Error)
	}
	entity := ToDomain(model)
	return &entity, nil
}

func (r *Postgres) ExistByNetworkId(ctx context.Context, networkId string) (bool, error) {
	var exists bool
	err := r.getDB(ctx).
		Model(&faucetwallet.FaucetWallet{}).
		Select("count(*) > 0").
		Where("network_id = ?", networkId).
		Find(&exists).
		Error
	if err != nil {
		return false, fmt.Errorf("failed to check existence of FaucetWallet for network ID %s: %w", networkId, err)
	}
	return exists, nil
}
