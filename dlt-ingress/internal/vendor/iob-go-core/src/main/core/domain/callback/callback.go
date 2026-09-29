package callback

import (
	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/basemodel"
)

type CallbackStatus = string

const (
	Pending CallbackStatus = "PENDING"
	Sent    CallbackStatus = "SENT"
	Failed  CallbackStatus = "FAILED"
)

type Callback struct {
	basemodel.Model
	ProcessId uuid.UUID      `gorm:"type:uuid;index"`
	Url       string         `gorm:"type:text;not null"`
	Status    CallbackStatus `gorm:"type:varchar(20);not null"`
	Attempts  uint
	LastError *string
}

type CallbackPayload struct {
	ProcessId uuid.UUID  `json:"processId"`
	Status    string     `json:"status"`
	EntityId  *uuid.UUID `json:"entityId,omitempty"`
}

type CallbackWithProcess struct {
	Callback
	EntityId      uuid.UUID `gorm:"column:entity_id"`
	ProcessStatus string    `gorm:"column:status"`
}

type CallbackConfigurator func(callback *Callback)

func New(config CallbackConfigurator) *Callback {
	callback := &Callback{
		Status:    Pending,
		Attempts:  0,
		LastError: nil,
	}
	config(callback)
	return callback
}
