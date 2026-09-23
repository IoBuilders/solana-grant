package eventstorepersistence

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SolanaTransactionAccount is one entry in a SolanaTransactionEvent's full
// account list, as returned by the RPC response for the transaction message
// — distinct from a SolanaInstructionAccount, which only lists the accounts
// a single instruction references.
type SolanaTransactionAccount struct {
	ID                 uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	TransactionEventID uuid.UUID `gorm:"column:transaction_event_id;type:uuid;not null;uniqueIndex:idx_solana_transaction_accounts_transaction_event_account"`

	// Index is the account's position in the transaction message's account
	// list, which is what gives it its meaning: SQL rows have no inherent
	// order to recover it from.
	Index int `gorm:"column:account_index;not null;uniqueIndex:idx_solana_transaction_accounts_transaction_event_account"`

	Address   string    `gorm:"column:address;type:varchar(64);not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (m *SolanaTransactionAccount) BeforeCreate(*gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
