package eventstore

import (
	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/basemodel"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/validate"
)

type Status string

const (
	Pending    Status = "PENDING"
	Processing Status = "PROCESSING"
	Succeeded  Status = "SUCCEEDED"
	Failed     Status = "FAILED"
)

var validTypes = map[Status]struct{}{
	Pending:    {},
	Processing: {},
	Succeeded:  {},
	Failed:     {},
}

func (t Status) IsValid() bool {
	_, ok := validTypes[t]
	return ok
}

func (t Status) IsFailed() bool {
	return t == Failed
}

type EventConsumer struct {
	basemodel.Model
	Type         string    `gorm:"type:varchar(255);not null"`
	Status       Status    `gorm:"type:varchar(10);not null"`
	EventStoreId uuid.UUID `gorm:"type:uuid;not null"`
	EventStore   EventStore
	IsCross      bool    `gorm:"type:boolean;not null;default:false"`
	ErrorDetails *string `gorm:"type:varchar(5000)"`
}

func NewEventConsumer(consumerType string, status string, isCross bool) (*EventConsumer, error) {
	if err := validate.StringMaxLength(consumerType, "Type", "EventConsumer", 255); err != nil {
		return nil, err
	}

	if !Status(status).IsValid() {
		return nil, domainerrors.NewEventConsumerInvalidStatusFailureDomainError(status)
	}

	return &EventConsumer{
		Type:    consumerType,
		Status:  Status(status),
		IsCross: isCross,
	}, nil
}

func (ec *EventConsumer) TransitToPending() error {
	if !ec.Status.IsFailed() {
		return domainerrors.NewEntityUnexpectedStatusDomainError("EventConsumer", ec.Id, string(ec.Status), string(Failed))
	}
	ec.Status = Pending
	return nil
}
