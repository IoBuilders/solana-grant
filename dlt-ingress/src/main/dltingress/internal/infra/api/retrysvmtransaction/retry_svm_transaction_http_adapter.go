package retrysvmtransaction

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/app/command/retrysvmtransaction"
	"dlt-ingress/src/main/dltingress/internal/infra/api/model"
	"net/http"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/command"
)

const UrlPath = "/admin/transactions/svm/retry"

type Endpoint struct {
	commandBus command.Bus
}

func NewEndpoint(commandBus command.Bus) *Endpoint {
	return &Endpoint{commandBus: commandBus}
}

// RetrySvmTransaction godoc
//
// @Summary      Retry an existing SVM transaction
// @Description  Retries a previously submitted Solana transaction that is stored in the system and has not landed on-chain yet.
// @Description
// @Description  The request must include the signature of the original transaction. A fresh blockhash is always
// @Description  fetched and the transaction re-signed; the rest of the instructions are carried over unchanged.
// @Description
// @Description  Compute unit parameters are optional:
// @Description  - `cuLimit` overrides the compute unit limit. If omitted, the original transaction's value is reused.
// @Description  - `cuPrice` overrides the compute unit price (priority fee). If omitted, the original value is reused.
// @Description  - If `useMultiplier=true`, the system applies the configured retry multiplier to the original compute unit price.
// @Tags         Admin
// @ID           retry-svm-transaction
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body     body      retrysvmtransaction.Request  true  "Retry SVM transaction request"
// @Success      200      {object}  model.RetryTransactionModel  "Retry transaction accepted successfully"
// @Failure      400      {object}  api.ErrorResponse
// @Failure      401      {object}  api.ErrorResponse
// @Failure      403      {object}  api.ErrorResponse
// @Failure      404      {object}  api.ErrorResponse
// @Failure      409      {object}  api.ErrorResponse
// @Failure      500      {object}  api.ErrorResponse
// @Router       /admin/transactions/svm/retry [post]
func (e *Endpoint) Handle(c context.Context, req *api.HttpAdapterRequest[Request]) (*api.HttpAdapterResponse[model.RetryTransactionModel], error) {
	response, err := e.commandBus.Dispatch(c, &retrysvmtransaction.Command{
		TransactionHash: req.Body.TxHash,
		CuLimit:         req.Body.CuLimit,
		CuPrice:         req.Body.CuPrice,
		UseMultiplier:   req.Body.UseMultiplier,
	})

	if err != nil {
		return nil, err
	}

	eventResponse := response.(*retrysvmtransaction.Response)

	return api.NewHttpAdapterResponse(http.StatusOK, &model.RetryTransactionModel{
		NewTxHash: eventResponse.NewTxId,
	}), nil
}
