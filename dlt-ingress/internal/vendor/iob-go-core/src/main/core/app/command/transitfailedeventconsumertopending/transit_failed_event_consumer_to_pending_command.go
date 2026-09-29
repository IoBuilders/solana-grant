package transitfailedeventconsumertopending

import (
	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/domain/eventstore/events"
)

type Command struct {
	EventConsumerId uuid.UUID
}

type Response struct {
	eventstoreevents.TransitFailedEventConsumerToPendingEvent
}
