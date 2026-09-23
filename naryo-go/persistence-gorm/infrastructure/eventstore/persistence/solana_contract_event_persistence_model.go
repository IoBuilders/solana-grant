package eventstorepersistence

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SolanaContractEvent struct {
	ID         uuid.UUID                `gorm:"column:id;type:uuid;primaryKey"`
	NodeID     uuid.UUID                `gorm:"column:node_id;type:uuid;not null"`
	EventName  string                   `gorm:"type:varchar(255);not null"`
	Status     string                   `gorm:"type:varchar(20);not null"`
	ProgramID  string                   `gorm:"column:program_id;type:varchar(44);not null"`
	Signature  string                   `gorm:"column:signature;type:varchar(2000);not null"`
	Slot       uint64                   `gorm:"column:slot;type:bigint;not null"`
	Parameters []ContractEventParameter `gorm:"foreignKey:ContractEventID;references:ID;constraint:OnDelete:CASCADE"`
	CreatedAt  time.Time                `gorm:"not null"`
	UpdatedAt  time.Time                `gorm:"not null"`
}

func (m *SolanaContractEvent) BeforeCreate(*gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
