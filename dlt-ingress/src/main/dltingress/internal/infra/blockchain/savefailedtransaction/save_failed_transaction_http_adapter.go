package savefailedtransaction

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/app/command/savefailedtransaction"
	"net/http"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/command"
)

const UrlPath = "/internal/dltingress/tx-revert"

type Endpoint struct {
	commandBus command.Bus
}

func NewEndpoint(commandBus command.Bus) *Endpoint {
	return &Endpoint{
		commandBus: commandBus,
	}
}

func (e *Endpoint) Listen(c context.Context, req *api.HttpAdapterRequest[Request]) (*api.HttpAdapterResponse[api.EmptyResponse], error) {
	_, err := e.commandBus.Dispatch(c, &savefailedtransaction.Command{
		TxId:         req.Body.TransactionHash,
		NetworkId:    req.Body.NetworkId,
		ErrorDetails: req.Body.RevertReason,
	})
	if err != nil {
		return nil, err
	}
	return api.NewEmptyHttpAdapterResponse(http.StatusOK), nil
}
