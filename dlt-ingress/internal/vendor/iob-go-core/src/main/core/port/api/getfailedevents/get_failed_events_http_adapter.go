package getfailedevents

import (
	"context"
	"fmt"
	"net/http"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api/pagination"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/app/query/getfailedevents"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/api/model"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/query"
)

func UrlPath(bcName string) string {
	return fmt.Sprintf("/admin/%s/failures/events", bcName)
}

type Endpoint struct {
	queryBus query.Bus
}

func NewEndpoint(queryBus query.Bus) *Endpoint {
	return &Endpoint{queryBus: queryBus}
}

func (e *Endpoint) Handle(ctx context.Context, req *api.HttpAdapterRequest[any]) (*api.HttpAdapterResponse[pagination.PageResponse], error) {
	params, err := req.GetPaginationParams()
	if err != nil {
		return nil, err
	}

	resp, err := query.Ask(ctx, e.queryBus, getfailedevents.Query{
		PaginationParams: params,
	})
	if err != nil {
		return nil, err
	}

	return api.NewHttpAdapterResponse(http.StatusOK, pagination.MapAndBuildResponse(resp, func(queryResponse *getfailedevents.EventConsumerQueryResponse) *model.EventConsumerResponse {
		return &model.EventConsumerResponse{
			Id:           queryResponse.Id,
			Type:         queryResponse.Type,
			Consumer:     queryResponse.Consumer,
			Payload:      queryResponse.Payload,
			CreatedAt:    queryResponse.CreatedAt,
			ErrorDetails: queryResponse.ErrorDetails,
		}
	})), nil
}
