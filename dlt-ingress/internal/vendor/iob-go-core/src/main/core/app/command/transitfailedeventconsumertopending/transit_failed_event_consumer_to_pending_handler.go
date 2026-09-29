package transitfailedeventconsumertopending

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/domain/eventstore/events"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/port/repository/eventstore"
)

type Handler struct {
	eventBus                event.Bus
	eventConsumerRepository eventstorerepo.EventConsumerRepository
}

func NewHandler(
	eventBus event.Bus,
	eventConsumerRepository eventstorerepo.EventConsumerRepository,
) *Handler {
	return &Handler{
		eventBus:                eventBus,
		eventConsumerRepository: eventConsumerRepository,
	}
}

func (h *Handler) Handle(ctx context.Context, cmd *Command) (*Response, error) {
	eventConsumer, err := h.eventConsumerRepository.FindById(ctx, cmd.EventConsumerId)
	if err != nil {
		return nil, err
	}

	if err = eventConsumer.TransitToPending(); err != nil {
		return nil, err
	}

	err = h.eventConsumerRepository.Save(ctx, eventConsumer)
	if err != nil {
		return nil, err
	}

	evt := eventstoreevents.TransitFailedEventConsumerToPendingEvent{
		BaseEvent:       *event.NewBaseEvent(),
		EventConsumerId: eventConsumer.Id,
	}
	err = h.eventBus.Publish(ctx, evt)
	if err != nil {
		return nil, err
	}

	return &Response{evt}, nil
}
