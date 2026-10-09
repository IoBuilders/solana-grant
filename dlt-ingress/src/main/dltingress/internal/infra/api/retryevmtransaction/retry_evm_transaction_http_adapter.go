package retryevmtransaction

import (
	"context"
	"net/http"

	"dlt-ingress/src/main/dltingress/internal/app/command/retryevmtransaction"
	"dlt-ingress/src/main/dltingress/internal/infra/api/model"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/command"
)

const UrlPath = "/admin/transactions/evm/retry"

type Endpoint struct {
	commandBus command.Bus
}

func NewEndpoint(commandBus command.Bus) *Endpoint {
	return &Endpoint{commandBus: commandBus}
}

// RetryEvmTransaction godoc
//
// @Summary      Retry an existing EVM transaction
// @Description  Retries a previously submitted transaction that is stored in the system and has not been executed yet on the target DLT.
// @Description
// @Description  The request must include the hash of the original transaction.
// @Description
// @Description  Fee parameters are optional:
// @Description  - For legacy transactions (type 0), `gasPrice` may be provided.
// @Description  - For EIP-1559 transactions (type 2), `maxPriorityFeePerGas` may be provided.
// @Description  - If a fee field is omitted, the value from the original transaction is reused.
// @Description  - If `useMultiplier=true`, the system applies the configured retry multiplier to the fee values.
// @Description
// @Description  The nonce is also optional. If omitted, the current address nonce is retrieved.
// @Tags         Admin
// @ID           retry-evm-transaction
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body     body      retryevmtransaction.Request  true  "Retry transaction request"
// @Success      200      {object}  model.RetryTransactionModel  "Retry transaction accepted successfully"
// @Failure      400      {object}  api.ErrorResponse
// @Failure      401      {object}  api.ErrorResponse
// @Failure      403      {object}  api.ErrorResponse
// @Failure      404      {object}  api.ErrorResponse
// @Failure      409      {object}  api.ErrorResponse
// @Failure      500      {object}  api.ErrorResponse
// @Router       /admin/transactions/evm/retry [post]
func (e *Endpoint) Handle(c context.Context, req *api.HttpAdapterRequest[Request]) (*api.HttpAdapterResponse[model.RetryTransactionModel], error) {
	response, err := e.commandBus.Dispatch(c, &retryevmtransaction.Command{
		TransactionHash:      req.Body.TxHash,
		GasLimit:             req.Body.GasLimit,
		GasPrice:             req.Body.GasPrice,
		MaxPriorityFeePerGas: req.Body.MaxPriorityFeePerGas,
		UseMultiplier:        req.Body.UseMultiplier,
		Nonce:                req.Body.Nonce,
	})

	if err != nil {
		return nil, err
	}

	eventResponse := response.(*retryevmtransaction.Response)

	return api.NewHttpAdapterResponse(http.StatusOK, &model.RetryTransactionModel{
		NewTxHash: eventResponse.NewTxId,
	}), nil
}
