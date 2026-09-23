package transactiongasestimator

import (
	"context"
	"dlt-ingress/src/main/dltingress/port/transactiongasestimator"

	"github.com/stretchr/testify/mock"
)

type TransactionGasEstimatorMock struct {
	mock.Mock
}

func (m *TransactionGasEstimatorMock) EstimateGas(ctx context.Context, request transactiongasestimator.EstimationRequest) (*transactiongasestimator.EstimationResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*transactiongasestimator.EstimationResponse), args.Error(1)
}

var _ transactiongasestimator.Port = (*TransactionGasEstimatorMock)(nil)
