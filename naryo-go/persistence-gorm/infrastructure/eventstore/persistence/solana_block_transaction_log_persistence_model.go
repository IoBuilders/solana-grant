package eventstorepersistence

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SolanaBlockTransactionLog struct {
	ID                 uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	BlockTransactionID uuid.UUID `gorm:"column:block_transaction_id;type:uuid;not null;uniqueIndex:idx_solana_block_transaction_logs_block_transaction_log"`
	Index              int       `gorm:"column:log_index;not null;uniqueIndex:idx_solana_block_transaction_logs_block_transaction_log"`
	Message            string    `gorm:"column:message;type:text;not null"`
	CreatedAt          time.Time `gorm:"not null"`
	UpdatedAt          time.Time `gorm:"not null"`
}

func (SolanaBlockTransactionLog) TableName() string {
	return "solana_block_transaction_logs"
}

func (m *SolanaBlockTransactionLog) BeforeCreate(*gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
