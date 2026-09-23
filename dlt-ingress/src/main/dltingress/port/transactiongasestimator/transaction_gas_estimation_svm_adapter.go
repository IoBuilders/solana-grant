package transactiongasestimator

import (
	"context"
	"fmt"
	"math/big"

	"dlt-ingress/src/main/dltingress/port/svm"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
)

type SvmTransactionGasEstimator struct {
	svmClientRegistry svm.ClientRegistry
}

func NewSvmTransactionGasEstimator(svmClientRegistry svm.ClientRegistry) *SvmTransactionGasEstimator {
	return &SvmTransactionGasEstimator{svmClientRegistry: svmClientRegistry}
}

func (s *SvmTransactionGasEstimator) EstimateGas(ctx context.Context, request EstimationRequest) (*EstimationResponse, error) {
	client, err := s.svmClientRegistry.GetClientForNetworkId(request.NetworkId)
	if err != nil {
		return nil, err
	}

	unitsConsumed, err := client.SimulateTransaction(ctx, request.Transaction.SVMTransactionResponse.SerializedTransaction)
	if err != nil {
		return nil, fmt.Errorf("error simulating SVM transaction on network %s: %w", request.NetworkId, err)
	}

	estimation, err := amount.New(new(big.Int).SetUint64(unitsConsumed), 0)
	if err != nil {
		return nil, fmt.Errorf("error converting compute units to amount: %w", err)
	}

	return &EstimationResponse{Estimation: estimation}, nil
}

var _ Port = (*SvmTransactionGasEstimator)(nil)
