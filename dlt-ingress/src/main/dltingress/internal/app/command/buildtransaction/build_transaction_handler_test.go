//go:build test

package buildtransaction

import (
	"context"
	"strings"
	"testing"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/app/service/txservice"
	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"dlt-ingress/src/main/dltingress/internal/infra/contracttransactionbuilder"
	ctbmock "dlt-ingress/src/main/dltingress/internal/infra/contracttransactionbuilder/mock"
	mock2 "dlt-ingress/src/main/dltingress/internal/infra/evm/mock"
	"dlt-ingress/src/main/dltingress/internal/infra/nonceprovider"
	npmock "dlt-ingress/src/main/dltingress/internal/infra/nonceprovider/mock"
	"dlt-ingress/src/main/dltingress/internal/infra/portcommon"
	svmmocks "dlt-ingress/src/main/dltingress/internal/infra/svm"
	"dlt-ingress/src/main/dltingress/internal/infra/transactiongasestimator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
)

const (
	validNetworkId            = "mainnet"
	validDlt                  = "EVM"
	validUrl                  = "http://localhost:8545"
	validChainId              = "1"
	validGasLimit             = "21000"
	validGasPrice             = "1000000000"
	validMaxPriorityFeePerGas = "1000000000"
	validFeeMultiplier        = "1.2"
	validMaxTxPoolSize        = 100
	validSenderDltAccountId   = "0x123"
	validSmartContractId      = "0x456"
	validSmartContractName    = "ERC20"
	validMethodName           = "transfer"
	validTransactionType      = 0
)

func init() {
	typedChainId, _ := amount.NewFromString(validChainId)
	typedGasLimit, _ := amount.NewFromString(validGasLimit)
	typedGasPrice, _ := amount.NewFromString(validGasPrice)
	typedMaxPriorityFeePerGas, _ := amount.NewFromString(validMaxPriorityFeePerGas)
	typedFeeMultiplier, _ := amount.NewFromString(validFeeMultiplier)

	config.DltIngressConfig = &config.Config{}
	config.DltIngressConfig.DltIngress.Networks = []*config.NetworkConfig{
		{
			Id:                   validNetworkId,
			Url:                  validUrl,
			Dlt:                  validDlt,
			ChainId:              *typedChainId,
			TransactionType:      validTransactionType,
			GasLimit:             *typedGasLimit,
			GasPrice:             *typedGasPrice,
			MaxPriorityFeePerGas: *typedMaxPriorityFeePerGas,
			FeeMultiplier:        *typedFeeMultiplier,
			MaxTxPoolSize:        validMaxTxPoolSize,
			GasLimitMultiplier:   *amount.Zero(),
		},
	}
}

func newHandler() (
	*CommandHandler,
	*npmock.NonceProviderMock,
	*contracttransactionbuilder.Registry,
	*mock2.EvmClientRegistryMock,
) {
	nonceProvider := new(npmock.NonceProviderMock)
	blockhashProvider := new(svmmocks.BlockhashProviderMock)
	registry := contracttransactionbuilder.NewRegistry()
	ethClientRegistry := new(mock2.EvmClientRegistryMock)
	svmClientRegistry := new(svmmocks.SvmClientRegistryMock)
	transactionGasEstimatorRegistry := transactiongasestimator.NewRegistry()

	txSvc := txservice.NewAppService(nonceProvider, blockhashProvider, *registry, ethClientRegistry, svmClientRegistry, *transactionGasEstimatorRegistry)
	handler := NewCommandHandler(txSvc)

	return handler, nonceProvider, registry, ethClientRegistry
}

func TestDltIngressBuildTransactionCommandHandler_Handle_Success(t *testing.T) {
	handler, nonceProvider, registry, ethClientRegistry := newHandler()

	ctx := context.Background()
	cmd := &Command{
		SenderDltAccountId: validSenderDltAccountId,
		SmartContractId:    validSmartContractId,
		SmartContractName:  validSmartContractName,
		MethodName:         validMethodName,
		MethodArgs:         map[string]any{},
		NetworkId:          validNetworkId,
	}

	nonce, _ := amount.NewFromString("0")
	nonceProvider.On("GetNonce", ctx, &nonceprovider.GetNonceRequest{
		NetworkId:    validNetworkId,
		DltAccountId: validSenderDltAccountId,
	}).Return(nonce, nil)

	networkGasPrice, _ := amount.NewFromString(validGasPrice)
	ethClient := new(mock2.EvmClientMock)
	ethClientRegistry.On("GetClientForNetworkId", ctx, validNetworkId).Return(ethClient, nil)
	ethClient.On("GetGasPrice", ctx).Return(networkGasPrice, nil)

	builder := new(ctbmock.ContractTransactionBuilderMock)
	registry.Register(common.EVM, validSmartContractName, builder)

	chainId, _ := amount.NewFromString(validChainId)
	gasLimit, _ := amount.NewFromString(validGasLimit)
	gasPrice, _ := amount.NewFromString(validGasPrice)
	value := amount.Zero()
	txResponse := &portcommon.TransactionResponse{
		EVMTransactionResponse: &portcommon.EVMTransactionResponse{
			Data:            "0x",
			To:              validSmartContractId,
			Nonce:           nonce,
			ChainId:         chainId,
			TransactionType: portcommon.TransactionTypeLegacy,
			GasLimit:        gasLimit,
			GasPrice:        gasPrice,
			Value:           value,
		},
	}
	builder.On("BuildTransaction", mock.Anything).Return(txResponse, nil)

	resp, err := handler.Handle(ctx, cmd)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, validSenderDltAccountId, resp.DltAccountId)
	assert.True(t, strings.HasPrefix(resp.Payload, "0x"), "payload should be RLP-encoded hex")

	nonceProvider.AssertNotCalled(t, "SetNonce", mock.Anything, mock.Anything)
	nonceProvider.AssertExpectations(t)
	builder.AssertExpectations(t)
	ethClientRegistry.AssertExpectations(t)
	ethClient.AssertExpectations(t)
}

func TestDltIngressBuildTransactionCommandHandler_Handle_NetworkNotFound(t *testing.T) {
	handler, _, _, _ := newHandler()

	resp, err := handler.Handle(context.Background(), &Command{
		NetworkId: "invalid-network",
	})

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, "Entity Network with invalid-network not found", err.Error())
}

func TestDltIngressBuildTransactionCommandHandler_Handle_BuildTransactionError(t *testing.T) {
	handler, nonceProvider, registry, ethClientRegistry := newHandler()

	ctx := context.Background()
	cmd := &Command{
		SenderDltAccountId: validSenderDltAccountId,
		SmartContractId:    validSmartContractId,
		SmartContractName:  validSmartContractName,
		MethodName:         validMethodName,
		MethodArgs:         map[string]any{},
		NetworkId:          validNetworkId,
	}

	nonce, _ := amount.NewFromString("0")
	nonceProvider.On("GetNonce", ctx, mock.Anything).Return(nonce, nil)

	networkGasPrice, _ := amount.NewFromString(validGasPrice)
	ethClient := new(mock2.EvmClientMock)
	ethClientRegistry.On("GetClientForNetworkId", ctx, validNetworkId).Return(ethClient, nil)
	ethClient.On("GetGasPrice", ctx).Return(networkGasPrice, nil)

	builder := new(ctbmock.ContractTransactionBuilderMock)
	registry.Register(common.EVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(nil, assert.AnError)

	resp, err := handler.Handle(ctx, cmd)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, assert.AnError)
}
