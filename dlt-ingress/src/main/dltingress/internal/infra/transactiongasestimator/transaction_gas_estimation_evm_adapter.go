package transactiongasestimator

import (
	"context"
	"fmt"

	"dlt-ingress/src/main/dltingress/internal/infra/evm"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
)

type EvmTransactionGasEstimator struct {
	evmClientRegistry evm.ClientRegistry
}

func NewEvmTransactionGasEstimator(evmClientRegistry evm.ClientRegistry) *EvmTransactionGasEstimator {
	return &EvmTransactionGasEstimator{
		evmClientRegistry,
	}
}

func (s *EvmTransactionGasEstimator) EstimateGas(ctx context.Context, request EstimationRequest) (*EstimationResponse, error) {

	ethClient, err := s.evmClientRegistry.GetClientForNetworkId(ctx, request.NetworkId)
	if err != nil {
		return nil, err
	}

	var result hexutil.Big

	data := map[string]any{
		"from":  request.From,
		"to":    request.Transaction.To,
		"data":  request.Transaction.Data,
		"value": hexutil.EncodeBig(request.Transaction.ValueBigInt()),
	}

	if err := ethClient.CallRpcMethod(ctx, &result, "eth_estimateGas", data); err != nil {
		return nil, fmt.Errorf("error sending gas estimation transaction to network: %w", err)
	}

	estimationAmount, err := amount.New(result.ToInt(), 0)

	if err != nil {
		return nil, fmt.Errorf("invalid estimation amount: %w", err)
	}

	return &EstimationResponse{Estimation: estimationAmount}, nil
}

var _ Port = (*EvmTransactionGasEstimator)(nil)
