package getfaucetwallet

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/app/query/getfaucetwallet"
	"dlt-ingress/src/main/dltingress/internal/infra/api/model"
	"net/http"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/query"
)

const UrlPath = "/admin/networks/:networkId/faucet-wallet"

type Endpoint struct {
	queryBus query.Bus
}

func NewEndpoint(queryBus query.Bus) *Endpoint {
	return &Endpoint{queryBus: queryBus}
}

// GetFaucetWallet godoc
//
// @Summary      Get a faucet wallet
// @Description  Retrieves the faucet wallet for the given network, including its current on-chain balance.
// @Tags         Admin
// @ID           get-faucet-wallet
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        networkId  path      string  true  "Network ID"
// @Success      200        {object}  model.FaucetWalletBalanceModel
// @Failure      400        {object}  api.ErrorResponse
// @Failure      401        {object}  api.ErrorResponse
// @Failure      404        {object}  api.ErrorResponse
// @Failure      500        {object}  api.ErrorResponse
// @Router       /admin/networks/{networkId}/faucet-wallet [get]
func (e *Endpoint) Handle(c context.Context, req *api.HttpAdapterRequest[any]) (*api.HttpAdapterResponse[model.FaucetWalletBalanceModel], error) {
	resp, err := query.Ask(c, e.queryBus, getfaucetwallet.Query{
		NetworkId: req.GetParam("networkId"),
	})
	if err != nil {
		return nil, err
	}

	return api.NewHttpAdapterResponse(http.StatusOK, &model.FaucetWalletBalanceModel{
		Id:               resp.FaucetWalletId.String(),
		NetworkId:        resp.NetworkId,
		DltAccountId:     resp.DltAccountId,
		FundingAmount:    resp.FundingAmount,
		BalanceThreshold: resp.BalanceThreshold,
		Balance:          resp.Balance,
		Enabled:          resp.Enabled,
	}), nil
}
