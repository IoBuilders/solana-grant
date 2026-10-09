package baserepo

import "github.com/google/uuid"

type UnsignedTransactionProcess struct {
	Id           uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProcessId    uuid.UUID `gorm:"type:uuid;not null"`
	DltAccountId string    `gorm:"type:varchar(255);not null"`
	Payload      string    `gorm:"type:text;not null"`
}
