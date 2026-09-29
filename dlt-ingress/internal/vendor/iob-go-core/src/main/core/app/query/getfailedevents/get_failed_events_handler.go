package getfailedevents

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/domain/eventstore"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/port/repository/eventstore"
)

type QueryHandler struct {
	eventStoreRepository eventstorerepo.EventConsumerRepository
}

func NewHandler(eventStoreRepository eventstorerepo.EventConsumerRepository) *QueryHandler {
	return &QueryHandler{eventStoreRepository: eventStoreRepository}
}

func (h *QueryHandler) Execute(ctx context.Context, query Query) (Response, error) {
	entities, err := h.eventStoreRepository.FindByStatusPaginated(ctx, eventstore.Failed, query.PaginationParams)
	if err != nil {
		return Response{}, err
	}

	total, err := h.eventStoreRepository.CountAllByFilters(ctx, eventstore.Failed)
	if err != nil {
		return Response{}, err
	}

	items := make([]*EventConsumerQueryResponse, len(entities))
	for i, n := range entities {
		items[i] = toItem(n)
	}

	return Response{
		Items:            items,
		TotalElements:    total,
		PaginationParams: query.PaginationParams,
	}, nil
}

func toItem(ec eventstore.EventConsumer) *EventConsumerQueryResponse {
	return &EventConsumerQueryResponse{
		Id:           ec.Id,
		Type:         ec.EventStore.Type,
		Consumer:     ec.Type,
		Payload:      ec.EventStore.Payload.String(),
		CreatedAt:    ec.CreatedAt,
		ErrorDetails: ec.ErrorDetails,
	}
}
