package eventstorepersistence

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SolanaBlockInstruction struct {
	ID                 uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	BlockTransactionID uuid.UUID `gorm:"column:block_transaction_id;type:uuid;not null;uniqueIndex:idx_solana_block_instructions_block_transaction_instruction"`
	// Index is not presentation order. In Solana an instruction is referenced by
	// its position within the transaction, so it is part of the natural key and
	// has to be stored: SQL rows have no inherent order to recover it from.
	Index     int    `gorm:"column:instruction_index;not null;uniqueIndex:idx_solana_block_instructions_block_transaction_instruction"`
	ProgramID string `gorm:"column:program_id;type:varchar(128);not null"`
	Data      []byte `gorm:"column:data;type:bytea"`
	Inner     bool   `gorm:"column:inner;not null"`

	Accounts  []SolanaBlockInstructionAccount `gorm:"foreignKey:BlockInstructionID;references:ID;constraint:OnDelete:CASCADE"`
	CreatedAt time.Time                       `gorm:"not null"`
	UpdatedAt time.Time                       `gorm:"not null"`
}

func (m *SolanaBlockInstruction) BeforeCreate(*gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
