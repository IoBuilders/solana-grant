package retryfailedevent

import (
	"context"
	"fmt"
	"net/http"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/app/command/transitfailedeventconsumertopending"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/command"
)

func UrlPath(bcName string) string {
	return fmt.Sprintf("/admin/%s/failures/events/:eventId/retry", bcName)
}

type Endpoint struct {
	commandBus command.Bus
}

func NewEndpoint(commandBus command.Bus) *Endpoint {
	return &Endpoint{commandBus: commandBus}
}

func (e *Endpoint) Handle(ctx context.Context, req *api.HttpAdapterRequest[any]) (*api.HttpAdapterResponse[api.EmptyResponse], error) {
	eventId, err := req.GetUUIDParam("eventId")
	if err != nil {
		return nil, err
	}

	_, err = e.commandBus.Dispatch(ctx, &transitfailedeventconsumertopending.Command{
		EventConsumerId: eventId,
	})
	if err != nil {
		return nil, err
	}

	return api.NewEmptyHttpAdapterResponse(http.StatusOK), nil
}
