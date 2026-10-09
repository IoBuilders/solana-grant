package createfaucetwallet

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/app/command/createfaucetwallet"
	"dlt-ingress/src/main/dltingress/internal/infra/api/model"
	"net/http"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/command"
)

const UrlPath = "/admin/networks/:networkId/faucet-wallet"

type Endpoint struct {
	commandBus command.Bus
}

func NewEndpoint(commandBus command.Bus) *Endpoint {
	return &Endpoint{commandBus: commandBus}
}

// CreateFaucetWallet godoc
//
// @Summary      Create a faucet wallet
// @Description  Creates a custody key and a faucet wallet for the given network.
// @Description
// @Description  fundingAmount and balanceThreshold are expressed in the network's smallest native unit (e.g. wei for EVM networks).
// @Tags         Admin
// @ID           create-faucet-wallet
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        networkId  path      string  true  "Network ID"
// @Param        body       body      createfaucetwallet.Request  true  "Create faucet wallet request"
// @Success      200        {object}  model.FaucetWalletModel  "Faucet wallet created successfully"
// @Failure      400        {object}  api.ErrorResponse
// @Failure      401        {object}  api.ErrorResponse
// @Failure      403        {object}  api.ErrorResponse
// @Failure      404        {object}  api.ErrorResponse
// @Failure      409        {object}  api.ErrorResponse
// @Failure      500        {object}  api.ErrorResponse
// @Router       /admin/networks/{networkId}/faucet-wallet [post]
func (e *Endpoint) Handle(c context.Context, req *api.HttpAdapterRequest[Request]) (*api.HttpAdapterResponse[model.FaucetWalletModel], error) {
	response, err := e.commandBus.Dispatch(c, &createfaucetwallet.Command{
		NetworkId:        req.GetParam("networkId"),
		FundingAmount:    req.Body.FundingAmount,
		BalanceThreshold: req.Body.BalanceThreshold,
	})
	if err != nil {
		return nil, err
	}

	cmdResponse := response.(*createfaucetwallet.Response)

	return api.NewHttpAdapterResponse(http.StatusOK, &model.FaucetWalletModel{
		Id:               cmdResponse.FaucetWalletId.String(),
		NetworkId:        cmdResponse.NetworkId,
		DltAccountId:     cmdResponse.DltAccountId,
		FundingAmount:    cmdResponse.FundingAmount,
		BalanceThreshold: cmdResponse.BalanceThreshold,
		Enabled:          cmdResponse.Enabled,
	}), nil
}
