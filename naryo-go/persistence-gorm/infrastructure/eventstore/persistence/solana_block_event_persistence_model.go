package eventstorepersistence

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SolanaBlockEvent struct {
	ID           uuid.UUID                `gorm:"column:id;type:uuid;primaryKey"`
	NodeID       uuid.UUID                `gorm:"column:node_id;type:uuid;not null;uniqueIndex:idx_solana_block_events_node_slot"`
	Slot         uint64                   `gorm:"type:bigint;not null;uniqueIndex:idx_solana_block_events_node_slot"`
	Blockhash    string                   `gorm:"type:varchar(44);not null"`
	BlockTime    *int64                   `gorm:"type:bigint"`
	Transactions []SolanaBlockTransaction `gorm:"foreignKey:BlockEventID;references:ID;constraint:OnDelete:CASCADE"`
	CreatedAt    time.Time                `gorm:"not null"`
	UpdatedAt    time.Time                `gorm:"not null"`
}

func (m *SolanaBlockEvent) BeforeCreate(*gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
