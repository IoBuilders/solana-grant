//go:build test

package evm_test

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"testing"

	"dlt-ingress/src/main/dltingress/internal/infra/evm"
	evmmocks "dlt-ingress/src/main/dltingress/internal/infra/evm/mock"
	"dlt-ingress/src/main/dltingress/internal/infra/noderatelimit"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/ratelimit"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const testAccount = "0x0000000000000000000000000000000000000001"

func TestDltIngressRateLimitedClient_CallRpcMethod_DelegatesWhenAllowed(t *testing.T) {
	inner := &evmmocks.EvmClientMock{}
	inner.On("CallRpcMethod", mock.Anything, mock.Anything, "eth_call", mock.Anything).Return(nil).Once()
	client := evm.NewRateLimitedClient(inner, ratelimit.New(ratelimit.WithBucket("TIER_1", 1, ratelimit.Methods("eth_call"))), nil)

	var result string
	err := client.CallRpcMethod(context.Background(), &result, "eth_call")

	require.NoError(t, err)
	inner.AssertExpectations(t)
}

func TestDltIngressRateLimitedClient_CallRpcMethod_RejectsWhenBucketExhausted(t *testing.T) {
	inner := &evmmocks.EvmClientMock{}
	inner.On("CallRpcMethod", mock.Anything, mock.Anything, "eth_call", mock.Anything).Return(nil).Once()
	client := evm.NewRateLimitedClient(inner, ratelimit.New(ratelimit.WithBucket("TIER_1", 1, ratelimit.Methods("eth_call"))), nil)

	var result string
	require.NoError(t, client.CallRpcMethod(context.Background(), &result, "eth_call"))
	err := client.CallRpcMethod(context.Background(), &result, "eth_call")

	require.Error(t, err)
	assert.True(t, ratelimit.IsExceededError(err))
	inner.AssertNumberOfCalls(t, "CallRpcMethod", 1)
}

func TestDltIngressRateLimitedClient_CallRpcMethod_Live429ExhaustsBucket(t *testing.T) {
	inner := &evmmocks.EvmClientMock{}
	nodeErr := &noderatelimit.TooManyRequestsError{}
	inner.On("CallRpcMethod", mock.Anything, mock.Anything, "eth_call", mock.Anything).Return(nodeErr).Once()
	client := evm.NewRateLimitedClient(inner, ratelimit.New(ratelimit.WithBucket("TIER_1", 100, ratelimit.Methods("eth_call"))), nil)

	var result string
	err := client.CallRpcMethod(context.Background(), &result, "eth_call")
	require.True(t, ratelimit.IsExceededError(err), "expected the node's 429 as ratelimit.ExceededError, got %T: %v", err, err)

	err = client.CallRpcMethod(context.Background(), &result, "eth_call")

	assert.True(t, ratelimit.IsExceededError(err))
	inner.AssertNumberOfCalls(t, "CallRpcMethod", 1)
}

func TestDltIngressRateLimitedClient_CallRpcMethod_NonRateLimitErrorKeepsBucket(t *testing.T) {
	inner := &evmmocks.EvmClientMock{}
	inner.On("CallRpcMethod", mock.Anything, mock.Anything, "eth_call", mock.Anything).Return(errors.New("boom")).Once()
	inner.On("CallRpcMethod", mock.Anything, mock.Anything, "eth_call", mock.Anything).Return(nil).Once()
	client := evm.NewRateLimitedClient(inner, ratelimit.New(ratelimit.WithBucket("TIER_1", 2, ratelimit.Methods("eth_call"))), nil)

	var result string
	require.Error(t, client.CallRpcMethod(context.Background(), &result, "eth_call"))
	err := client.CallRpcMethod(context.Background(), &result, "eth_call")

	require.NoError(t, err)
	inner.AssertNumberOfCalls(t, "CallRpcMethod", 2)
}

func TestDltIngressRateLimitedClient_TypedMethods_UseTheirRpcMethodBucket(t *testing.T) {
	balance, _ := amount.New(big.NewInt(1), 0)
	inner := &evmmocks.EvmClientMock{}
	inner.On("PendingNonceAt", mock.Anything, testAccount).Return(uint64(7), nil).Once()
	inner.On("GetBaseFeePerGas", mock.Anything).Return(balance, nil).Once()
	inner.On("GetGasPrice", mock.Anything).Return(balance, nil).Once()
	inner.On("GetBalance", mock.Anything, testAccount).Return(balance, nil).Once()
	inner.On("HeaderByNumber", mock.Anything, mock.Anything).Return(&types.Header{}, nil).Once()
	inner.On("BlockByNumber", mock.Anything, mock.Anything).Return(&types.Block{}, nil).Once()
	inner.On("NetworkID", mock.Anything).Return(big.NewInt(1), nil).Once()
	inner.On("CallContract", mock.Anything, mock.Anything, mock.Anything).Return([]byte{}, nil).Once()
	// Several typed methods share a JSON-RPC method (and so a bucket), hence a fresh client per method.
	newClient := func() evm.Client {
		return evm.NewRateLimitedClient(inner, ratelimit.New(
			ratelimit.WithBucket("NONCE", 1, ratelimit.Methods("eth_getTransactionCount")),
			ratelimit.WithBucket("BLOCK", 1, ratelimit.Methods("eth_getBlockByNumber")),
			ratelimit.WithBucket("GAS_PRICE", 1, ratelimit.Methods("eth_gasPrice")),
			ratelimit.WithBucket("BALANCE", 1, ratelimit.Methods("eth_getBalance")),
			ratelimit.WithBucket("NET_VERSION", 1, ratelimit.Methods("net_version")),
			ratelimit.WithBucket("CALL", 1, ratelimit.Methods("eth_call")),
		), nil)
	}
	ctx := context.Background()

	calls := map[string]func(client evm.Client) error{
		"PendingNonceAt":   func(client evm.Client) error { _, err := client.PendingNonceAt(ctx, testAccount); return err },
		"GetBaseFeePerGas": func(client evm.Client) error { _, err := client.GetBaseFeePerGas(ctx); return err },
		"GetGasPrice":      func(client evm.Client) error { _, err := client.GetGasPrice(ctx); return err },
		"GetBalance":       func(client evm.Client) error { _, err := client.GetBalance(ctx, testAccount); return err },
		"HeaderByNumber":   func(client evm.Client) error { _, err := client.HeaderByNumber(ctx, nil); return err },
		"BlockByNumber":    func(client evm.Client) error { _, err := client.BlockByNumber(ctx, nil); return err },
		"NetworkID":        func(client evm.Client) error { _, err := client.NetworkID(ctx); return err },
		"CallContract": func(client evm.Client) error {
			_, err := client.CallContract(ctx, ethereum.CallMsg{}, nil)
			return err
		},
	}
	for name, call := range calls {
		client := newClient()
		require.NoError(t, call(client), name)
		err := call(client)
		assert.True(t, ratelimit.IsExceededError(err), fmt.Sprintf("%s second call should be rate limited", name))
	}
	inner.AssertExpectations(t)
}

func TestDltIngressRateLimitedClient_Close_Delegates(t *testing.T) {
	inner := &evmmocks.EvmClientMock{}
	inner.On("Close").Return().Once()
	client := evm.NewRateLimitedClient(inner, ratelimit.New(), nil)

	client.Close()

	inner.AssertExpectations(t)
}
