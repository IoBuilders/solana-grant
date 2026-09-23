package transanctionsender

import (
	"context"
	"dlt-ingress/src/main/dltingress/port/transanctionsender"
	ethmocks "dlt-ingress/src/test/dltingress/port/evm"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	evmNetworkId = "mainnet"
	evmDlt       = "EVM"
	evmSignedTx  = "0xf86c808504a817c800825208943535353535353535353535353535353535353535880de0b6b3a76400008025a0"
	evmTxId      = "0xabc123def456"
)

func TestDltIngressEvmTransactionSender_SendTransaction_Success(t *testing.T) {
	ethClientRegistry := new(ethmocks.EvmClientRegistryMock)
	ethClient := new(ethmocks.EvmClientMock)
	sender := transanctionsender.NewEvmTransactionSender(ethClientRegistry)

	ctx := context.Background()
	req := transanctionsender.SendTransactionRequest{
		SignedTransaction: evmSignedTx,
		Dlt:               evmDlt,
		NetworkId:         evmNetworkId,
	}

	ethClientRegistry.On("GetClientForNetworkId", ctx, evmNetworkId).Return(ethClient, nil)
	ethClient.On("CallRpcMethod", ctx, mock.AnythingOfType("*string"), "eth_sendRawTransaction", mock.Anything).
		Run(func(args mock.Arguments) {
			*args.Get(1).(*string) = evmTxId
		}).
		Return(nil)

	resp, err := sender.SendTransaction(ctx, req)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, evmTxId, resp.TxId)
	ethClientRegistry.AssertExpectations(t)
	ethClient.AssertExpectations(t)
}

func TestDltIngressEvmTransactionSender_SendTransaction_ClientRegistryError(t *testing.T) {
	ethClientRegistry := new(ethmocks.EvmClientRegistryMock)
	sender := transanctionsender.NewEvmTransactionSender(ethClientRegistry)

	ctx := context.Background()
	req := transanctionsender.SendTransactionRequest{
		SignedTransaction: evmSignedTx,
		Dlt:               evmDlt,
		NetworkId:         evmNetworkId,
	}

	ethClientRegistry.On("GetClientForNetworkId", ctx, evmNetworkId).Return(nil, assert.AnError)

	resp, err := sender.SendTransaction(ctx, req)

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, assert.AnError)
	ethClientRegistry.AssertExpectations(t)
}

func TestDltIngressEvmTransactionSender_SendTransaction_RpcError(t *testing.T) {
	ethClientRegistry := new(ethmocks.EvmClientRegistryMock)
	ethClient := new(ethmocks.EvmClientMock)
	sender := transanctionsender.NewEvmTransactionSender(ethClientRegistry)

	ctx := context.Background()
	req := transanctionsender.SendTransactionRequest{
		SignedTransaction: evmSignedTx,
		Dlt:               evmDlt,
		NetworkId:         evmNetworkId,
	}

	ethClientRegistry.On("GetClientForNetworkId", ctx, evmNetworkId).Return(ethClient, nil)
	ethClient.On("CallRpcMethod", ctx, mock.AnythingOfType("*string"), "eth_sendRawTransaction", mock.Anything).
		Return(assert.AnError)

	resp, err := sender.SendTransaction(ctx, req)

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, assert.AnError)
	assert.ErrorContains(t, err, "error sending transaction to network")
	ethClientRegistry.AssertExpectations(t)
	ethClient.AssertExpectations(t)
}
