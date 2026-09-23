package eventstorepersistence

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ContractEventParameter struct {
	ID              uuid.UUID      `gorm:"column:id;type:uuid;primaryKey"`
	Type            string         `gorm:"type:varchar(20);not null"`
	Position        int            `gorm:"type:int;not null"`
	Value           datatypes.JSON `gorm:"type:jsonb;not null"`
	ContractEventID uuid.UUID      `gorm:"not null"`
	CreatedAt       time.Time      `gorm:"not null"`
	UpdatedAt       time.Time      `gorm:"not null"`
}

func (m *ContractEventParameter) BeforeCreate(*gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
