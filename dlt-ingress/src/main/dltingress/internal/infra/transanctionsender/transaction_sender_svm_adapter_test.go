package transanctionsender

import (
	"context"
	svmmocks "dlt-ingress/src/main/dltingress/internal/infra/svm"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	svmNetworkId = "solana-mainnet"
	svmDlt       = "SVM"
	svmSignedTx  = "AQABAgIDBAUGBwgJ"
	svmTxId      = "5VfYmGBL5dV1aH6dvB1Yd9k3oQ8Wd2nQ2Yk9eQ7Wd2nQ2Yk9eQ7Wd2nQ2Yk9eQ7Wd2"
)

func TestDltIngressSvmTransactionSender_SendTransaction_Success(t *testing.T) {
	svmClientRegistry := new(svmmocks.SvmClientRegistryMock)
	svmClient := new(svmmocks.SvmClientMock)
	sender := NewSvmTransactionSender(svmClientRegistry)

	ctx := context.Background()
	req := SendTransactionRequest{
		SignedTransaction: svmSignedTx,
		Dlt:               svmDlt,
		NetworkId:         svmNetworkId,
	}

	svmClientRegistry.On("GetClientForNetworkId", svmNetworkId).Return(svmClient, nil)
	svmClient.On("SendTransaction", ctx, svmSignedTx).Return(svmTxId, nil)

	resp, err := sender.SendTransaction(ctx, req)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, svmTxId, resp.TxId)
	svmClientRegistry.AssertExpectations(t)
	svmClient.AssertExpectations(t)
}

func TestDltIngressSvmTransactionSender_SendTransaction_ClientRegistryError(t *testing.T) {
	svmClientRegistry := new(svmmocks.SvmClientRegistryMock)
	sender := NewSvmTransactionSender(svmClientRegistry)

	ctx := context.Background()
	req := SendTransactionRequest{
		SignedTransaction: svmSignedTx,
		Dlt:               svmDlt,
		NetworkId:         svmNetworkId,
	}

	svmClientRegistry.On("GetClientForNetworkId", svmNetworkId).Return(nil, assert.AnError)

	resp, err := sender.SendTransaction(ctx, req)

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, assert.AnError)
	svmClientRegistry.AssertExpectations(t)
}

func TestDltIngressSvmTransactionSender_SendTransaction_ClienttError(t *testing.T) {
	svmClientRegistry := new(svmmocks.SvmClientRegistryMock)
	svmClient := new(svmmocks.SvmClientMock)
	sender := NewSvmTransactionSender(svmClientRegistry)

	ctx := context.Background()
	req := SendTransactionRequest{
		SignedTransaction: svmSignedTx,
		Dlt:               svmDlt,
		NetworkId:         svmNetworkId,
	}

	svmClientRegistry.On("GetClientForNetworkId", svmNetworkId).Return(svmClient, nil)
	svmClient.On("SendTransaction", ctx, svmSignedTx).Return("", assert.AnError)

	resp, err := sender.SendTransaction(ctx, req)

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, assert.AnError)
	assert.ErrorContains(t, err, "error sending transaction to network")
	svmClientRegistry.AssertExpectations(t)
	svmClient.AssertExpectations(t)
}
