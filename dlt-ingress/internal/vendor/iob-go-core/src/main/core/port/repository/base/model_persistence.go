package baserepo

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Model struct {
	Id        uuid.UUID `gorm:"type:uuid;primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (e *Model) BeforeCreate(tx *gorm.DB) error {
	if e.Id == uuid.Nil {
		e.Id = uuid.New()
	}
	e.CreatedAt = time.Now()
	return nil
}
