package eventstorepersistence

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SolanaBlockTransaction struct {
	ID            uuid.UUID                       `gorm:"column:id;type:uuid;primaryKey"`
	BlockEventID  uuid.UUID                       `gorm:"column:block_event_id;type:uuid;not null"`
	NodeID        uuid.UUID                       `gorm:"column:node_id;type:uuid;not null;uniqueIndex:idx_solana_block_transactions_node_transaction"`
	TransactionID string                          `gorm:"column:transaction_id;type:varchar(88);not null;uniqueIndex:idx_solana_block_transactions_node_transaction"`
	Err           *string                         `gorm:"column:err;type:varchar(5000)"`
	Slot          uint64                          `gorm:"column:slot;not null;index"`
	Instructions  []SolanaBlockInstruction        `gorm:"foreignKey:BlockTransactionID;references:ID;constraint:OnDelete:CASCADE"`
	Accounts      []SolanaBlockTransactionAccount `gorm:"foreignKey:BlockTransactionID;references:ID;constraint:OnDelete:CASCADE"`
	Logs          []SolanaBlockTransactionLog     `gorm:"foreignKey:BlockTransactionID;references:ID;constraint:OnDelete:CASCADE"`
	CreatedAt     time.Time                       `gorm:"not null"`
	UpdatedAt     time.Time                       `gorm:"not null"`
}

func (m *SolanaBlockTransaction) BeforeCreate(*gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
