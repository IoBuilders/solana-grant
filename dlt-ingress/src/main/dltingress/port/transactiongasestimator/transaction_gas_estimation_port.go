package transactiongasestimator

import (
	"context"
	"dlt-ingress/src/main/dltingress/port/portcommon"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
)

type Port interface {
	EstimateGas(ctx context.Context, request EstimationRequest) (*EstimationResponse, error)
}
type EstimationRequest struct {
	NetworkId   string
	From        string
	Transaction *portcommon.TransactionResponse
}
type EstimationResponse struct {
	Estimation *amount.Amount
}
