package txqueueslotrepo

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/db"
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

func (r *Postgres) DeleteExpired(ctx context.Context) error {
	result := r.getDB(ctx).Unscoped().Where("expires_at < ?", time.Now()).Delete(&TxQueueSlot{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete expired TxQueueSlot: %w", result.Error)
	}

	return nil
}

func (r *Postgres) DeleteFirstByNetworkId(ctx context.Context, networkId string) error {
	var slot TxQueueSlot
	result := r.getDB(ctx).
		Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Where("network_id = ? AND expires_at > ?", networkId, time.Now()).
		Order("created_at").
		Limit(1).
		Find(&slot)
	if result.Error != nil {
		return fmt.Errorf("failed to find first TxQueueSlot for network id %s: %w", networkId, result.Error)
	}
	if result.RowsAffected == 0 {
		return nil
	}

	if err := r.getDB(ctx).Unscoped().Where("id = ?", slot.Id).Delete(&TxQueueSlot{}).Error; err != nil {
		return fmt.Errorf("failed to delete first TxQueueSlot for network id %s: %w", networkId, err)
	}

	return nil
}

func (r *Postgres) CountByNetworkIdAndNotExpired(ctx context.Context, networkId string) (int, error) {
	var count int64
	result := r.getDB(ctx).Model(&TxQueueSlot{}).Where("network_id = ? AND expires_at > ?", networkId, time.Now()).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count TxQueueSlot entities for network id %s: %w", networkId, result.Error)
	}
	return int(count), nil
}
