package updatefaucetwallet

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/app/command/updatefaucetwallet"
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

// UpdateFaucetWallet godoc
//
// @Summary      Update a faucet wallet
// @Description  Updates the funding amount and balance threshold of the faucet wallet for the given network.
// @Description
// @Description  fundingAmount and balanceThreshold are expressed in the network's smallest native unit (e.g. wei for EVM networks). enabled toggles whether the faucet wallet is used to fund accounts.
// @Tags         Admin
// @ID           update-faucet-wallet
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        networkId  path      string  true  "Network ID"
// @Param        body       body      updatefaucetwallet.Request  true  "Update faucet wallet request"
// @Success      200        {object}  model.FaucetWalletModel  "Faucet wallet updated successfully"
// @Failure      400        {object}  api.ErrorResponse
// @Failure      401        {object}  api.ErrorResponse
// @Failure      403        {object}  api.ErrorResponse
// @Failure      404        {object}  api.ErrorResponse
// @Failure      500        {object}  api.ErrorResponse
// @Router       /admin/networks/{networkId}/faucet-wallet [put]
func (e *Endpoint) Handle(c context.Context, req *api.HttpAdapterRequest[Request]) (*api.HttpAdapterResponse[model.FaucetWalletModel], error) {
	response, err := e.commandBus.Dispatch(c, &updatefaucetwallet.Command{
		NetworkId:        req.GetParam("networkId"),
		FundingAmount:    req.Body.FundingAmount,
		BalanceThreshold: req.Body.BalanceThreshold,
		Enabled:          *req.Body.Enabled,
	})
	if err != nil {
		return nil, err
	}

	cmdResponse := response.(*updatefaucetwallet.Response)

	return api.NewHttpAdapterResponse(http.StatusOK, &model.FaucetWalletModel{
		Id:               cmdResponse.FaucetWalletId.String(),
		NetworkId:        cmdResponse.NetworkId,
		FundingAmount:    cmdResponse.FundingAmount,
		BalanceThreshold: cmdResponse.BalanceThreshold,
		Enabled:          cmdResponse.Enabled,
	}), nil
}
