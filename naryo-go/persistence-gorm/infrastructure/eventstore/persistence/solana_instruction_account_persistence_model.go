package eventstorepersistence

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SolanaInstructionAccount struct {
	ID            uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	InstructionID uuid.UUID `gorm:"column:instruction_id;type:uuid;not null;uniqueIndex:idx_solana_instruction_accounts_instruction_account"`

	// Index is the account's position in the instruction, which is what gives it
	// its meaning: which account a program reads as the payer, the mint or the
	// authority is decided by position, not by name.
	Index int `gorm:"column:account_index;not null;uniqueIndex:idx_solana_instruction_accounts_instruction_account"`

	Address   string    `gorm:"column:address;type:varchar(64);not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (m *SolanaInstructionAccount) BeforeCreate(*gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
