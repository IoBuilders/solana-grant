package eventstorepersistence

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SolanaLatestBlock struct {
	ID        uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	NodeID    uuid.UUID `gorm:"column:node_id;type:uuid;not null;uniqueIndex:idx_solana_latest_block_node"`
	Slot      uint64    `gorm:"type:bigint;not null;"`
	Blockhash string    `gorm:"type:varchar(44);not null"`
	BlockTime *int64    `gorm:"type:bigint"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (m *SolanaLatestBlock) BeforeCreate(*gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
