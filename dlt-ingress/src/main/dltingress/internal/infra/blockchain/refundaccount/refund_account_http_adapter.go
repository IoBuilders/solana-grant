package refundaccount

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/app/service/fundaccount"
	"net/http"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api"
)

const UrlPath = "/internal/dltingress/refund-account"

type Endpoint struct {
	appService fundaccount.AppServiceInterface
}

func NewEndpoint(appService fundaccount.AppServiceInterface) *Endpoint {
	return &Endpoint{appService: appService}
}

func (e *Endpoint) Listen(ctx context.Context, req *api.HttpAdapterRequest[Request]) (*api.HttpAdapterResponse[api.EmptyResponse], error) {
	err := e.appService.Execute(ctx, &fundaccount.Request{
		NetworkId:    &req.Body.NetworkId,
		DltAccountId: req.Body.Signer,
	})
	if err != nil {
		return nil, err
	}
	return api.NewEmptyHttpAdapterResponse(http.StatusOK), nil
}
