package txqueueslotrepo

import (
	"context"
	"dlt-ingress/src/main/dltingress/domain/txqueueslot"
	"fmt"
	"time"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/db"
	"gorm.io/gorm"
)

type PostgresTxQueueSlotRepository struct {
	db *gorm.DB
	tm db.TransactionManager
}

func NewPostgresTxQueueSlotRepository(gormDB *gorm.DB, tm db.TransactionManager) *PostgresTxQueueSlotRepository {
	return &PostgresTxQueueSlotRepository{db: gormDB, tm: tm}
}

func (r *PostgresTxQueueSlotRepository) DeleteExpired(ctx context.Context) error {
	result := r.getDB(ctx).Unscoped().Where("expires_at < ?", time.Now()).Delete(&txqueueslot.TxQueueSlot{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete expired TxQueueSlot: %w", result.Error)
	}

	return nil
}

func (r *PostgresTxQueueSlotRepository) DeleteFirstByNetworkId(ctx context.Context, networkId string) error {
	result := r.getDB(ctx).Unscoped().Where("network_id = ?", networkId).Order("created_at").Limit(1).Delete(&txqueueslot.TxQueueSlot{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete first TxQueueSlot for network id %s: %w", networkId, result.Error)
	}

	return nil
}

func (r *PostgresTxQueueSlotRepository) CountByNetworkIdAndNotExpired(ctx context.Context, networkId string) (int, error) {
	var count int64
	result := r.getDB(ctx).Model(&txqueueslot.TxQueueSlot{}).Where("network_id = ? AND expires_at > ?", networkId, time.Now()).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count TxQueueSlot entities for network id %s: %w", networkId, result.Error)
	}
	return int(count), nil
}
