//go:build test

package contractcaller

import (
	"context"
	"dlt-ingress/src/main/dltingress/port/contractcaller"
	ethmocks "dlt-ingress/src/test/dltingress/port/evm"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	callerNetworkId  = "mainnet"
	callerContractTo = "0x92BDb699940f5a5Fb21C7D6EFA5a67C84B84A085"
	callerCallData   = "0xca221a58"
	callerResultData = "0x000000000000000000000000000000000000000000000000000000000000002a"
)

func TestDltIngressEvmContractCaller_Call_Success(t *testing.T) {
	registry := new(ethmocks.EvmClientRegistryMock)
	client := new(ethmocks.EvmClientMock)
	caller := contractcaller.NewEvmContractCaller(registry)

	ctx := context.Background()
	req := contractcaller.CallRequest{
		To:        callerContractTo,
		Data:      callerCallData,
		NetworkId: callerNetworkId,
	}

	registry.On("GetClientForNetworkId", ctx, callerNetworkId).Return(client, nil)
	client.On("CallRpcMethod", ctx, mock.AnythingOfType("*string"), "eth_call", mock.Anything).
		Run(func(args mock.Arguments) {
			*args.Get(1).(*string) = callerResultData
		}).
		Return(nil)

	resp, err := caller.Call(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, callerResultData, resp.Data)
	registry.AssertExpectations(t)
	client.AssertExpectations(t)
}

func TestDltIngressEvmContractCaller_Call_ClientRegistryError(t *testing.T) {
	registry := new(ethmocks.EvmClientRegistryMock)
	caller := contractcaller.NewEvmContractCaller(registry)

	ctx := context.Background()
	req := contractcaller.CallRequest{
		To:        callerContractTo,
		Data:      callerCallData,
		NetworkId: callerNetworkId,
	}

	registry.On("GetClientForNetworkId", ctx, callerNetworkId).Return(nil, assert.AnError)

	resp, err := caller.Call(ctx, req)

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, assert.AnError)
	registry.AssertExpectations(t)
}

func TestDltIngressEvmContractCaller_Call_RpcError(t *testing.T) {
	registry := new(ethmocks.EvmClientRegistryMock)
	client := new(ethmocks.EvmClientMock)
	caller := contractcaller.NewEvmContractCaller(registry)

	ctx := context.Background()
	req := contractcaller.CallRequest{
		To:        callerContractTo,
		Data:      callerCallData,
		NetworkId: callerNetworkId,
	}

	registry.On("GetClientForNetworkId", ctx, callerNetworkId).Return(client, nil)
	client.On("CallRpcMethod", ctx, mock.AnythingOfType("*string"), "eth_call", mock.Anything).
		Return(assert.AnError)

	resp, err := caller.Call(ctx, req)

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, assert.AnError)
	assert.ErrorContains(t, err, "error calling eth_call on contract")
	registry.AssertExpectations(t)
	client.AssertExpectations(t)
}

func TestDltIngressEvmContractCaller_ImplementsPortInterface(t *testing.T) {
	var _ contractcaller.Port = (*contractcaller.EvmContractCaller)(nil)
}
