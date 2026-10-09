package transactiongasestimator

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"dlt-ingress/src/main/dltingress/internal/infra/portcommon"
	"dlt-ingress/src/main/dltingress/internal/infra/svm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AppService decides whether to refetch the blockhash with errors.Is(err, svm.ErrBlockhashNotFound) on what
// EstimateGas returns, so this adapter must keep the sentinel reachable through its own wrapping.
func TestDltIngressSvmTransactionGasEstimator_EstimateGas_PreservesBlockhashNotFoundSentinel(t *testing.T) {
	const networkId = "solana-mainnet"
	const serializedTx = "AQABAgME"

	registry := new(svm.SvmClientRegistryMock)
	client := new(svm.SvmClientMock)
	registry.On("GetClientForNetworkId", networkId).Return(client, nil)
	client.On("SimulateTransaction", context.Background(), serializedTx).
		Return(uint64(0), fmt.Errorf("%w: BlockhashNotFound", svm.ErrBlockhashNotFound))

	estimator := NewSvmTransactionGasEstimator(registry)

	_, err := estimator.EstimateGas(context.Background(), EstimationRequest{
		NetworkId: networkId,
		Transaction: &portcommon.TransactionResponse{
			SVMTransactionResponse: &portcommon.SVMTransactionResponse{SerializedTransaction: serializedTx},
		},
	})

	require.Error(t, err)
	assert.True(t, errors.Is(err, svm.ErrBlockhashNotFound), "expected ErrBlockhashNotFound to survive the adapter's wrapping, got %v", err)
	client.AssertExpectations(t)
}
