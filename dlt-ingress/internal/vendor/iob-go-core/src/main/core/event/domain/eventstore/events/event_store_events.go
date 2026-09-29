package eventstoreevents

import (
	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"
)

type TransitFailedEventConsumerToPendingEvent struct {
	event.BaseEvent
	EventConsumerId uuid.UUID
}
