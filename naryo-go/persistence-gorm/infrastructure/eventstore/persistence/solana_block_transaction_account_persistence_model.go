package eventstorepersistence

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SolanaBlockTransactionAccount is one entry in a SolanaBlockTransaction's
// full account list, as returned by the RPC response for the transaction
// message — distinct from a SolanaBlockInstructionAccount, which only lists
// the accounts a single instruction references.
type SolanaBlockTransactionAccount struct {
	ID                 uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	BlockTransactionID uuid.UUID `gorm:"column:block_transaction_id;type:uuid;not null;uniqueIndex:idx_solana_block_transaction_accounts_block_transaction_account"`

	// Index is the account's position in the transaction message's account
	// list, which is what gives it its meaning: SQL rows have no inherent
	// order to recover it from.
	Index int `gorm:"column:account_index;not null;uniqueIndex:idx_solana_block_transaction_accounts_block_transaction_account"`

	Address   string    `gorm:"column:address;type:varchar(64);not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (m *SolanaBlockTransactionAccount) BeforeCreate(*gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
