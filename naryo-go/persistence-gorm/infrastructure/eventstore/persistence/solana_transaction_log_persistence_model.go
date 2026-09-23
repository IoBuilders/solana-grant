package eventstorepersistence

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SolanaTransactionLog struct {
	ID                 uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	TransactionEventID uuid.UUID `gorm:"column:transaction_event_id;type:uuid;not null;uniqueIndex:idx_solana_transaction_logs_transaction_event_log"`
	Index              int       `gorm:"column:log_index;not null;uniqueIndex:idx_solana_transaction_logs_transaction_event_log"`
	Message            string    `gorm:"column:message;type:text;not null"`
	CreatedAt          time.Time `gorm:"not null"`
	UpdatedAt          time.Time `gorm:"not null"`
}

func (SolanaTransactionLog) TableName() string {
	return "solana_transaction_logs"
}

func (m *SolanaTransactionLog) BeforeCreate(*gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
