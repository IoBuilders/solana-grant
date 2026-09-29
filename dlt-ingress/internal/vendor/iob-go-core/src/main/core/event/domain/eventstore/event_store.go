package eventstore

import (
	"encoding/json"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/basemodel"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/validate"
	"gorm.io/datatypes"
)

type PublicationType string

const (
	Memory PublicationType = "MEMORY"
	HTTP   PublicationType = "HTTP"
)

var validPublicationTypes = map[PublicationType]struct{}{
	Memory: {},
	HTTP:   {},
}

func (t PublicationType) IsValid() bool {
	_, ok := validPublicationTypes[t]
	return ok
}

type EventStore struct {
	basemodel.Model
	Type            string          `gorm:"type:varchar(255);not null"`
	Payload         datatypes.JSON  `gorm:"type:jsonb;not null;index:idx_event_stores_tx_hash_ordered,expression:((payload->>'TransactionHash')::text),where:deleted_at IS NULL"`
	EventConsumers  []EventConsumer `gorm:"foreignKey:EventStoreId"`
	TraceParent     string          `gorm:"type:varchar(55);not null"`
	PublicationType PublicationType `gorm:"type:varchar(10);not null;default:MEMORY"`
	IsCross         bool            `gorm:"type:boolean;not null;default:false"`
}

func NewEventStore(eventType string, payload datatypes.JSON, eventConsumers []EventConsumer, traceParent string, publicationType string, isCross bool) (*EventStore, error) {
	if err := validate.StringMaxLength(eventType, "Type", "EventStore", 255); err != nil {
		return nil, err
	}

	var unmarshalledPayload map[string]interface{}
	err := json.Unmarshal(payload, &unmarshalledPayload)
	if err != nil || len(unmarshalledPayload) == 0 {
		return nil, domainerrors.NewEventStoreEmptyPayloadDomainError()
	}

	if !PublicationType(publicationType).IsValid() {
		return nil, domainerrors.NewEventStoreInvalidPublicationTypeDomainError(publicationType)
	}

	return &EventStore{
		Type:            eventType,
		Payload:         payload,
		EventConsumers:  eventConsumers,
		TraceParent:     traceParent,
		PublicationType: PublicationType(publicationType),
		IsCross:         isCross,
	}, nil
}
