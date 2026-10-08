package txservice

import (
	"context"
	"fmt"
	"testing"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"dlt-ingress/src/main/dltingress/internal/infra/contracttransactionbuilder"
	ctbmock "dlt-ingress/src/main/dltingress/internal/infra/contracttransactionbuilder/mock"
	mock2 "dlt-ingress/src/main/dltingress/internal/infra/evm/mock"
	"dlt-ingress/src/main/dltingress/internal/infra/nonceprovider"
	npmock "dlt-ingress/src/main/dltingress/internal/infra/nonceprovider/mock"
	"dlt-ingress/src/main/dltingress/internal/infra/portcommon"
	svmmocks "dlt-ingress/src/main/dltingress/internal/infra/svm"
	"dlt-ingress/src/main/dltingress/internal/infra/transactiongasestimator"
	gasmock "dlt-ingress/src/main/dltingress/internal/infra/transactiongasestimator/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
)

const (
	validNetworkId = "mainnet"
	validDlt       = "EVM"
	validUrl       = "http://localhost:8545"
	validChainId   = "1"
	// In NetworkConfig, gasLimit is both the transaction seed and the ceiling the engine
	// will not exceed, so the fixture carries the real Hedera ceiling.
	validGasLimit             = "15000000"
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

func newService() (
	*AppService,
	*npmock.NonceProviderMock,
	*contracttransactionbuilder.Registry,
	*mock2.EvmClientRegistryMock,
	*transactiongasestimator.Registry,
	*svmmocks.BlockhashProviderMock,
	*svmmocks.SvmClientRegistryMock,
) {
	nonceProvider := new(npmock.NonceProviderMock)
	blockhashProvider := new(svmmocks.BlockhashProviderMock)
	registry := contracttransactionbuilder.NewRegistry()
	ethClientRegistry := new(mock2.EvmClientRegistryMock)
	svmClientRegistry := new(svmmocks.SvmClientRegistryMock)
	gasEstimatorRegistry := transactiongasestimator.NewRegistry()
	svc := NewAppService(nonceProvider, blockhashProvider, *registry, ethClientRegistry, svmClientRegistry, *gasEstimatorRegistry)
	return svc, nonceProvider, registry, ethClientRegistry, gasEstimatorRegistry, blockhashProvider, svmClientRegistry
}

func validLegacyTxResponse(nonce *amount.Amount) *portcommon.TransactionResponse {
	gasLimit, _ := amount.NewFromString(validGasLimit)
	gasPrice, _ := amount.NewFromString(validGasPrice)
	return &portcommon.TransactionResponse{
		EVMTransactionResponse: &portcommon.EVMTransactionResponse{
			Data:            "0x",
			To:              validSmartContractId,
			Nonce:           nonce,
			TransactionType: portcommon.TransactionTypeLegacy,
			GasLimit:        gasLimit,
			GasPrice:        gasPrice,
		},
	}
}

func validRequest() TransactionRequest {
	return TransactionRequest{
		SenderDltAccountId: validSenderDltAccountId,
		SmartContractId:    validSmartContractId,
		SmartContractName:  validSmartContractName,
		MethodName:         validMethodName,
		MethodArgs:         map[string]any{},
		NetworkId:          validNetworkId,
	}
}

func TestDltIngressTxService_Evm_PrepareTransaction_Legacy_Success(t *testing.T) {
	svc, nonceProvider, txBuilderRegistry, ethClientRegistry, _, _, _ := newService()
	network := *config.DltIngressConfig.DltIngress.Networks[0]

	ctx := context.Background()
	nonce, _ := amount.NewFromString("0")
	nonceProvider.On("GetNonce", ctx, &nonceprovider.GetNonceRequest{
		NetworkId:    validNetworkId,
		DltAccountId: validSenderDltAccountId,
	}).Return(nonce, nil)

	networkGasPrice, _ := amount.NewFromString(validGasPrice)
	ethClient := new(mock2.EvmClientMock)
	ethClientRegistry.On("GetClientForNetworkId", ctx, validNetworkId).Return(ethClient, nil)
	ethClient.On("GetGasPrice", ctx).Return(networkGasPrice, nil)

	txResp := validLegacyTxResponse(nonce)
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.EVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(txResp, nil)

	result, err := svc.PrepareTransaction(ctx, validRequest(), network)

	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, txResp, result)
	assert.Equal(t, nonce, result.Nonce)
	nonceProvider.AssertExpectations(t)
	ethClientRegistry.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	builder.AssertExpectations(t)
}

func TestDltIngressTxService_Evm_PrepareTransaction_Legacy_GasPrice_EvmClientError_ReturnsError(t *testing.T) {
	svc, nonceProvider, txBuilderRegistry, ethClientRegistry, _, _, _ := newService()
	network := *config.DltIngressConfig.DltIngress.Networks[0]

	ctx := context.Background()
	nonce, _ := amount.NewFromString("0")
	nonceProvider.On("GetNonce", ctx, mock.Anything).Return(nonce, nil)

	ethClientRegistry.On("GetClientForNetworkId", ctx, validNetworkId).
		Return((*mock2.EvmClientMock)(nil), assert.AnError)

	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.EVM, validSmartContractName, builder)

	result, err := svc.PrepareTransaction(ctx, validRequest(), network)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, assert.AnError)
	ethClientRegistry.AssertExpectations(t)
	builder.AssertNotCalled(t, "BuildTransaction")
}

func TestDltIngressTxService_Evm_PrepareTransaction_Legacy_GasPrice_FetchError_FallsBackToConfigured(t *testing.T) {
	svc, nonceProvider, txBuilderRegistry, ethClientRegistry, _, _, _ := newService()
	network := *config.DltIngressConfig.DltIngress.Networks[0]

	ctx := context.Background()
	nonce, _ := amount.NewFromString("0")
	nonceProvider.On("GetNonce", ctx, mock.Anything).Return(nonce, nil)

	ethClient := new(mock2.EvmClientMock)
	ethClientRegistry.On("GetClientForNetworkId", ctx, validNetworkId).Return(ethClient, nil)
	ethClient.On("GetGasPrice", ctx).Return((*amount.Amount)(nil), assert.AnError)

	txResp := validLegacyTxResponse(nonce)
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.EVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(txResp, nil)

	result, err := svc.PrepareTransaction(ctx, validRequest(), network)

	assert.Nil(t, err)
	assert.NotNil(t, result)
	ethClientRegistry.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	builder.AssertExpectations(t)
}

func TestDltIngressTxService_Evm_PrepareTransaction_DynamicFee_Success(t *testing.T) {
	svc, nonceProvider, txBuilderRegistry, ethClientRegistry, _, _, _ := newService()

	typedChainId, _ := amount.NewFromString(validChainId)
	typedGasLimit, _ := amount.NewFromString(validGasLimit)
	typedMaxPriorityFeePerGas, _ := amount.NewFromString(validMaxPriorityFeePerGas)
	typedFeeMultiplier, _ := amount.NewFromString(validFeeMultiplier)
	network := config.NetworkConfig{
		Id:                   validNetworkId,
		Url:                  validUrl,
		Dlt:                  validDlt,
		ChainId:              *typedChainId,
		TransactionType:      portcommon.TransactionTypeDynamicFee,
		GasLimit:             *typedGasLimit,
		MaxPriorityFeePerGas: *typedMaxPriorityFeePerGas,
		FeeMultiplier:        *typedFeeMultiplier,
		MaxTxPoolSize:        validMaxTxPoolSize,
		GasLimitMultiplier:   *amount.Zero(),
	}

	ctx := context.Background()
	nonce, _ := amount.NewFromString("0")
	nonceProvider.On("GetNonce", ctx, mock.Anything).Return(nonce, nil)

	baseFee, _ := amount.NewFromString("1000000000")
	ethClient := new(mock2.EvmClientMock)
	ethClientRegistry.On("GetClientForNetworkId", ctx, validNetworkId).Return(ethClient, nil)
	ethClient.On("GetBaseFeePerGas", ctx).Return(baseFee, nil)

	incrementedBaseFee := baseFee.MulWithDecimals(network.FeeMultiplier, 0)
	expectedMaxFeePerGas := incrementedBaseFee.Add(network.MaxPriorityFeePerGas)

	gasLimit, _ := amount.NewFromString(validGasLimit)
	txResp := &portcommon.TransactionResponse{
		EVMTransactionResponse: &portcommon.EVMTransactionResponse{
			Data:                 "0x",
			To:                   validSmartContractId,
			Nonce:                nonce,
			TransactionType:      portcommon.TransactionTypeDynamicFee,
			GasLimit:             gasLimit,
			MaxPriorityFeePerGas: typedMaxPriorityFeePerGas,
			MaxFeePerGas:         &expectedMaxFeePerGas,
		},
	}
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.EVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.MatchedBy(func(req *contracttransactionbuilder.BuildTransactionRequest) bool {
		return req.EVMBuildTransactionRequest != nil &&
			req.EVMBuildTransactionRequest.MaxFeePerGas != nil &&
			req.EVMBuildTransactionRequest.MaxFeePerGas.Equal(expectedMaxFeePerGas)
	})).Return(txResp, nil)

	result, err := svc.PrepareTransaction(ctx, validRequest(), network)

	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.MaxFeePerGas)
	assert.True(t, result.MaxFeePerGas.Equal(expectedMaxFeePerGas))
	ethClientRegistry.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	builder.AssertExpectations(t)
}

func TestDltIngressTxService_Evm_PrepareTransaction_GetNonce_Error(t *testing.T) {
	svc, nonceProvider, txBuilderRegistry, _, _, _, _ := newService()
	network := *config.DltIngressConfig.DltIngress.Networks[0]

	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.EVM, validSmartContractName, builder)

	nonceProvider.On("GetNonce", mock.Anything, mock.Anything).
		Return((*amount.Amount)(nil), assert.AnError)

	result, err := svc.PrepareTransaction(context.Background(), validRequest(), network)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, assert.AnError)
	nonceProvider.AssertExpectations(t)
	builder.AssertNotCalled(t, "BuildTransaction")
}

func TestDltIngressTxService_Evm_PrepareTransaction_BuilderNotFound(t *testing.T) {
	svc, nonceProvider, _, _, _, _, _ := newService()
	network := *config.DltIngressConfig.DltIngress.Networks[0]

	nonce, _ := amount.NewFromString("0")
	nonceProvider.On("GetNonce", mock.Anything, mock.Anything).Return(nonce, nil)

	req := validRequest()
	req.SmartContractName = "UNKNOWN_CONTRACT"

	result, err := svc.PrepareTransaction(context.Background(), req, network)

	assert.Nil(t, result)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("contract transaction builder not found for dlt EVM and smart contract name %s", req.SmartContractName), err.Error())
}

func TestDltIngressTxService_Evm_PrepareTransaction_BuildTransaction_Error(t *testing.T) {
	svc, nonceProvider, txBuilderRegistry, ethClientRegistry, _, _, _ := newService()
	network := *config.DltIngressConfig.DltIngress.Networks[0]

	ctx := context.Background()
	nonce, _ := amount.NewFromString("0")
	nonceProvider.On("GetNonce", ctx, mock.Anything).Return(nonce, nil)

	networkGasPrice, _ := amount.NewFromString(validGasPrice)
	ethClient := new(mock2.EvmClientMock)
	ethClientRegistry.On("GetClientForNetworkId", ctx, validNetworkId).Return(ethClient, nil)
	ethClient.On("GetGasPrice", ctx).Return(networkGasPrice, nil)

	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.EVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(nil, assert.AnError)

	result, err := svc.PrepareTransaction(ctx, validRequest(), network)

	assert.Nil(t, result)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	ethClientRegistry.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	builder.AssertExpectations(t)
}

func TestDltIngressTxService_Evm_PrepareTransaction_DynamicFee_EvmClientError(t *testing.T) {
	svc, nonceProvider, txBuilderRegistry, ethClientRegistry, _, _, _ := newService()

	typedChainId, _ := amount.NewFromString(validChainId)
	typedGasLimit, _ := amount.NewFromString(validGasLimit)
	typedMaxPriorityFeePerGas, _ := amount.NewFromString(validMaxPriorityFeePerGas)
	typedFeeMultiplier, _ := amount.NewFromString(validFeeMultiplier)
	network := config.NetworkConfig{
		Id:                   validNetworkId,
		Dlt:                  validDlt,
		ChainId:              *typedChainId,
		TransactionType:      portcommon.TransactionTypeDynamicFee,
		GasLimit:             *typedGasLimit,
		MaxPriorityFeePerGas: *typedMaxPriorityFeePerGas,
		FeeMultiplier:        *typedFeeMultiplier,
		GasLimitMultiplier:   *amount.One(),
	}

	ctx := context.Background()
	nonce, _ := amount.NewFromString("0")
	nonceProvider.On("GetNonce", ctx, mock.Anything).Return(nonce, nil)

	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.EVM, validSmartContractName, builder)

	ethClientRegistry.On("GetClientForNetworkId", ctx, validNetworkId).
		Return((*mock2.EvmClientMock)(nil), assert.AnError)

	result, err := svc.PrepareTransaction(ctx, validRequest(), network)

	assert.Nil(t, result)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	ethClientRegistry.AssertExpectations(t)
	builder.AssertNotCalled(t, "BuildTransaction", mock.Anything)
}

func TestDltIngressTxService_Evm_PrepareTransaction_GasEstimation_Success(t *testing.T) {
	svc, nonceProvider, txBuilderRegistry, ethClientRegistry, gasEstimatorRegistry, _, _ := newService()
	network := *config.DltIngressConfig.DltIngress.Networks[0]
	network.GasLimitMultiplier = *amount.One()

	ctx := context.Background()
	nonce, _ := amount.NewFromString("0")
	nonceProvider.On("GetNonce", ctx, mock.Anything).Return(nonce, nil)

	networkGasPrice, _ := amount.NewFromString(validGasPrice)
	ethClient := new(mock2.EvmClientMock)
	ethClientRegistry.On("GetClientForNetworkId", ctx, validNetworkId).Return(ethClient, nil)
	ethClient.On("GetGasPrice", ctx).Return(networkGasPrice, nil)

	originalGasLimit, _ := amount.NewFromString(validGasLimit)
	txResp := validLegacyTxResponse(nonce)
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.EVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(txResp, nil)

	estimation, _ := amount.NewFromString("30000")
	estimator := new(gasmock.TransactionGasEstimatorMock)
	gasEstimatorRegistry.Register(common.EVM, estimator)
	estimator.On("EstimateGas", ctx, mock.MatchedBy(func(req transactiongasestimator.EstimationRequest) bool {
		return req.NetworkId == validNetworkId && req.From == validSenderDltAccountId
	})).Return(&transactiongasestimator.EstimationResponse{Estimation: estimation}, nil)

	expectedGasLimit := estimation.MulWithDecimals(network.GasLimitMultiplier, 0)
	overrideResp := validLegacyTxResponse(nonce)
	overrideResp.GasLimit = &expectedGasLimit
	builder.On("OverrideGasLimit", mock.Anything, mock.Anything, mock.MatchedBy(func(a *amount.Amount) bool {
		return a.Equal(expectedGasLimit)
	})).Return(overrideResp, nil).Once()

	result, err := svc.PrepareTransaction(ctx, validRequest(), network)

	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.GasLimit.Equal(*originalGasLimit))
	assert.True(t, result.GasLimit.Equal(expectedGasLimit))
	ethClientRegistry.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	estimator.AssertExpectations(t)
	builder.AssertExpectations(t)
}

func TestDltIngressTxService_Evm_PrepareTransaction_GasEstimation_Error_LogsWarning_ReturnsTransaction(t *testing.T) {
	svc, nonceProvider, txBuilderRegistry, ethClientRegistry, gasEstimatorRegistry, _, _ := newService()
	network := *config.DltIngressConfig.DltIngress.Networks[0]
	network.GasLimitMultiplier = *amount.One()

	ctx := context.Background()
	nonce, _ := amount.NewFromString("0")
	nonceProvider.On("GetNonce", ctx, mock.Anything).Return(nonce, nil)

	networkGasPrice, _ := amount.NewFromString(validGasPrice)
	ethClient := new(mock2.EvmClientMock)
	ethClientRegistry.On("GetClientForNetworkId", ctx, validNetworkId).Return(ethClient, nil)
	ethClient.On("GetGasPrice", ctx).Return(networkGasPrice, nil)

	txResp := validLegacyTxResponse(nonce)
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.EVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(txResp, nil)

	estimator := new(gasmock.TransactionGasEstimatorMock)
	gasEstimatorRegistry.Register(common.EVM, estimator)
	estimator.On("EstimateGas", ctx, mock.Anything).
		Return((*transactiongasestimator.EstimationResponse)(nil), fmt.Errorf("node unavailable"))

	result, err := svc.PrepareTransaction(ctx, validRequest(), network)

	assert.Nil(t, err)
	assert.Equal(t, txResp, result)
	ethClientRegistry.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	estimator.AssertExpectations(t)
	builder.AssertExpectations(t)
}

func TestDltIngressTxService_Evm_PrepareTransaction_GasEstimation_Skipped_WhenMultiplierBelowOne(t *testing.T) {
	svc, nonceProvider, txBuilderRegistry, ethClientRegistry, gasEstimatorRegistry, _, _ := newService()
	network := *config.DltIngressConfig.DltIngress.Networks[0]
	network.GasLimitMultiplier = *amount.Zero()

	ctx := context.Background()
	nonce, _ := amount.NewFromString("0")
	nonceProvider.On("GetNonce", ctx, mock.Anything).Return(nonce, nil)

	networkGasPrice, _ := amount.NewFromString(validGasPrice)
	ethClient := new(mock2.EvmClientMock)
	ethClientRegistry.On("GetClientForNetworkId", ctx, validNetworkId).Return(ethClient, nil)
	ethClient.On("GetGasPrice", ctx).Return(networkGasPrice, nil)

	txResp := validLegacyTxResponse(nonce)
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.EVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(txResp, nil)

	estimator := new(gasmock.TransactionGasEstimatorMock)
	gasEstimatorRegistry.Register(common.EVM, estimator)

	result, err := svc.PrepareTransaction(ctx, validRequest(), network)

	assert.Nil(t, err)
	assert.NotNil(t, result)
	estimator.AssertNotCalled(t, "EstimateGas", mock.Anything, mock.Anything)
	ethClientRegistry.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	builder.AssertExpectations(t)
}

func TestDltIngressTxService_Evm_PrepareTransaction_GasLimitBelowMax_PassesThrough(t *testing.T) {
	svc, nonceProvider, txBuilderRegistry, ethClientRegistry, gasEstimatorRegistry, _, _ := newService()
	network := *config.DltIngressConfig.DltIngress.Networks[0]
	network.GasLimitMultiplier = *amount.One()

	ctx := context.Background()
	nonce, _ := amount.NewFromString("0")
	nonceProvider.On("GetNonce", ctx, mock.Anything).Return(nonce, nil)

	networkGasPrice, _ := amount.NewFromString(validGasPrice)
	ethClient := new(mock2.EvmClientMock)
	ethClientRegistry.On("GetClientForNetworkId", ctx, validNetworkId).Return(ethClient, nil)
	ethClient.On("GetGasPrice", ctx).Return(networkGasPrice, nil)

	txResp := validLegacyTxResponse(nonce)
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.EVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(txResp, nil)

	// 14,400,000 sits just below the 15,000,000 ceiling, so it must be applied untouched.
	estimation, _ := amount.NewFromString("14400000")
	estimator := new(gasmock.TransactionGasEstimatorMock)
	gasEstimatorRegistry.Register(common.EVM, estimator)
	estimator.On("EstimateGas", ctx, mock.Anything).
		Return(&transactiongasestimator.EstimationResponse{Estimation: estimation}, nil)

	overrideResp := validLegacyTxResponse(nonce)
	overrideResp.GasLimit = estimation
	builder.On("OverrideGasLimit", mock.Anything, mock.Anything, mock.MatchedBy(func(a *amount.Amount) bool {
		return a.Equal(*estimation)
	})).Return(overrideResp, nil).Once()

	result, err := svc.PrepareTransaction(ctx, validRequest(), network)

	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint64(14400000), result.GasLimit.RawValue().Uint64())
	estimator.AssertExpectations(t)
	builder.AssertExpectations(t)
}

func TestDltIngressTxService_Evm_PrepareTransaction_GasLimitAboveMax_FailsFast(t *testing.T) {
	svc, nonceProvider, txBuilderRegistry, ethClientRegistry, gasEstimatorRegistry, _, _ := newService()
	network := *config.DltIngressConfig.DltIngress.Networks[0]
	network.GasLimitMultiplier = *amount.One()

	ctx := context.Background()
	nonce, _ := amount.NewFromString("0")
	nonceProvider.On("GetNonce", ctx, mock.Anything).Return(nonce, nil)

	networkGasPrice, _ := amount.NewFromString(validGasPrice)
	ethClient := new(mock2.EvmClientMock)
	ethClientRegistry.On("GetClientForNetworkId", ctx, validNetworkId).Return(ethClient, nil)
	ethClient.On("GetGasPrice", ctx).Return(networkGasPrice, nil)

	txResp := validLegacyTxResponse(nonce)
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.EVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(txResp, nil)

	// 15,600,000 exceeds the 15,000,000 ceiling: the transaction must be refused, not clamped.
	estimation, _ := amount.NewFromString("15600000")
	estimator := new(gasmock.TransactionGasEstimatorMock)
	gasEstimatorRegistry.Register(common.EVM, estimator)
	estimator.On("EstimateGas", ctx, mock.Anything).
		Return(&transactiongasestimator.EstimationResponse{Estimation: estimation}, nil)

	result, err := svc.PrepareTransaction(ctx, validRequest(), network)

	// The error must propagate rather than be degraded to a warning: swallowing it here would
	// let the un-overridden seed reach the signer.
	assert.Nil(t, result)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "15600000")
	assert.Contains(t, err.Error(), validGasLimit)
	builder.AssertNotCalled(t, "OverrideGasLimit")
	estimator.AssertExpectations(t)
}

func TestDltIngressTxService_Evm_PrepareTransaction_MultipliedGasLimitAboveMax_SkipsOverride(t *testing.T) {
	svc, nonceProvider, txBuilderRegistry, ethClientRegistry, gasEstimatorRegistry, _, _ := newService()
	network := *config.DltIngressConfig.DltIngress.Networks[0]
	multiplier, _ := amount.NewFromString("2")
	network.GasLimitMultiplier = *multiplier

	ctx := context.Background()
	nonce, _ := amount.NewFromString("0")
	nonceProvider.On("GetNonce", ctx, mock.Anything).Return(nonce, nil)

	networkGasPrice, _ := amount.NewFromString(validGasPrice)
	ethClient := new(mock2.EvmClientMock)
	ethClientRegistry.On("GetClientForNetworkId", ctx, validNetworkId).Return(ethClient, nil)
	ethClient.On("GetGasPrice", ctx).Return(networkGasPrice, nil)

	txResp := validLegacyTxResponse(nonce)
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.EVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(txResp, nil)

	// 10,000,000 sits below the 15,000,000 ceiling on its own, but doubled by the
	// multiplier it would land at 20,000,000, above the ceiling.
	estimation, _ := amount.NewFromString("10000000")
	estimator := new(gasmock.TransactionGasEstimatorMock)
	gasEstimatorRegistry.Register(common.EVM, estimator)
	estimator.On("EstimateGas", ctx, mock.Anything).
		Return(&transactiongasestimator.EstimationResponse{Estimation: estimation}, nil)

	result, err := svc.PrepareTransaction(ctx, validRequest(), network)

	// The raw estimate is within the cap, so this must not be treated as a failure: the
	// override is simply skipped and the original, un-overridden transaction goes through.
	assert.Nil(t, err)
	assert.Equal(t, txResp, result)
	builder.AssertNotCalled(t, "OverrideGasLimit")
	estimator.AssertExpectations(t)
	builder.AssertExpectations(t)
}

func TestDltIngressTxService_Evm_PrepareTransaction_GasEstimation_Error_SeedAtCeilingReachesSigner(t *testing.T) {
	svc, nonceProvider, txBuilderRegistry, ethClientRegistry, gasEstimatorRegistry, _, _ := newService()
	network := *config.DltIngressConfig.DltIngress.Networks[0]
	network.GasLimitMultiplier = *amount.One()

	ctx := context.Background()
	nonce, _ := amount.NewFromString("0")
	nonceProvider.On("GetNonce", ctx, mock.Anything).Return(nonce, nil)

	networkGasPrice, _ := amount.NewFromString(validGasPrice)
	ethClient := new(mock2.EvmClientMock)
	ethClientRegistry.On("GetClientForNetworkId", ctx, validNetworkId).Return(ethClient, nil)
	ethClient.On("GetGasPrice", ctx).Return(networkGasPrice, nil)

	txResp := validLegacyTxResponse(nonce)
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.EVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(txResp, nil)

	estimator := new(gasmock.TransactionGasEstimatorMock)
	gasEstimatorRegistry.Register(common.EVM, estimator)
	estimator.On("EstimateGas", ctx, mock.Anything).
		Return((*transactiongasestimator.EstimationResponse)(nil), fmt.Errorf("node unavailable"))

	result, err := svc.PrepareTransaction(ctx, validRequest(), network)

	// A failing estimation stays non-fatal, and the seed it leaves behind is already the ceiling,
	// so nothing above it can reach the signer and there is nothing to clamp.
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint64(15000000), result.GasLimit.RawValue().Uint64())
	builder.AssertNotCalled(t, "OverrideGasLimit")
	estimator.AssertExpectations(t)
	builder.AssertExpectations(t)
}

func TestDltIngressTxService_Evm_PrepareTransaction_MultiplierBelowOne_SeedAtCeilingReachesSigner(t *testing.T) {
	svc, nonceProvider, txBuilderRegistry, ethClientRegistry, gasEstimatorRegistry, _, _ := newService()
	network := *config.DltIngressConfig.DltIngress.Networks[0]
	network.GasLimitMultiplier = *amount.Zero()

	ctx := context.Background()
	nonce, _ := amount.NewFromString("0")
	nonceProvider.On("GetNonce", ctx, mock.Anything).Return(nonce, nil)

	networkGasPrice, _ := amount.NewFromString(validGasPrice)
	ethClient := new(mock2.EvmClientMock)
	ethClientRegistry.On("GetClientForNetworkId", ctx, validNetworkId).Return(ethClient, nil)
	ethClient.On("GetGasPrice", ctx).Return(networkGasPrice, nil)

	txResp := validLegacyTxResponse(nonce)
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.EVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(txResp, nil)

	estimator := new(gasmock.TransactionGasEstimatorMock)
	gasEstimatorRegistry.Register(common.EVM, estimator)

	result, err := svc.PrepareTransaction(ctx, validRequest(), network)

	// Estimation is skipped entirely, so the seed is what gets signed — and the seed is the ceiling.
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint64(15000000), result.GasLimit.RawValue().Uint64())
	estimator.AssertNotCalled(t, "EstimateGas", mock.Anything, mock.Anything)
	builder.AssertNotCalled(t, "OverrideGasLimit")
	builder.AssertExpectations(t)
}

func TestDltIngressTxService_Evm_PrepareTransaction_GasLimitUnset_NoCapApplied(t *testing.T) {
	svc, nonceProvider, txBuilderRegistry, ethClientRegistry, gasEstimatorRegistry, _, _ := newService()
	network := *config.DltIngressConfig.DltIngress.Networks[0]
	network.GasLimitMultiplier = *amount.One()
	network.GasLimit = amount.Amount{}

	ctx := context.Background()
	nonce, _ := amount.NewFromString("0")
	nonceProvider.On("GetNonce", ctx, mock.Anything).Return(nonce, nil)

	networkGasPrice, _ := amount.NewFromString(validGasPrice)
	ethClient := new(mock2.EvmClientMock)
	ethClientRegistry.On("GetClientForNetworkId", ctx, validNetworkId).Return(ethClient, nil)
	ethClient.On("GetGasPrice", ctx).Return(networkGasPrice, nil)

	txResp := validLegacyTxResponse(nonce)
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.EVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(txResp, nil)

	estimation, _ := amount.NewFromString("40000000")
	estimator := new(gasmock.TransactionGasEstimatorMock)
	gasEstimatorRegistry.Register(common.EVM, estimator)
	estimator.On("EstimateGas", ctx, mock.Anything).
		Return(&transactiongasestimator.EstimationResponse{Estimation: estimation}, nil)

	overrideResp := validLegacyTxResponse(nonce)
	overrideResp.GasLimit = estimation
	builder.On("OverrideGasLimit", mock.Anything, mock.Anything, mock.MatchedBy(func(a *amount.Amount) bool {
		return a.Equal(*estimation)
	})).Return(overrideResp, nil).Once()

	result, err := svc.PrepareTransaction(ctx, validRequest(), network)

	// An absent gasLimit means "no cap", never "cap at zero" — and comparing against an
	// uninitialised amount.Amount would panic, so the guard has to short-circuit first.
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint64(40000000), result.GasLimit.RawValue().Uint64())
	builder.AssertExpectations(t)
}

const (
	validSvmNetworkId  = "solana-mainnet"
	validSvmDlt        = "SVM"
	validSvmUrl        = "https://api.mainnet-beta.solana.com"
	validSvmBlockhash  = "GHtXQBsoZHVnNFa9YevAzFr17DJjgHXk3ycTKD5xD3Zi"
	validSvmSender     = "9xQeWvG816bUx9EPjHmaT23yvVM2ZWbrrpZb9PusVFin"
	validSvmContractId = "11111111111111111111111111111112"
)

func svmNetwork(gasLimitMultiplier *amount.Amount) config.NetworkConfig {
	maxCuPrice, _ := amount.NewFromString("1000000000")
	return config.NetworkConfig{
		Id:                 validSvmNetworkId,
		Url:                validSvmUrl,
		Dlt:                validSvmDlt,
		GasLimitMultiplier: *gasLimitMultiplier,
		MaxCuPrice:         *maxCuPrice,
	}
}

func validSvmTxResponse() *portcommon.TransactionResponse {
	return &portcommon.TransactionResponse{
		SVMTransactionResponse: &portcommon.SVMTransactionResponse{
			SerializedTransaction: "AQABAgIDBAUGBwgJ",
			FeePayer:              validSvmSender,
			RecentBlockhash:       validSvmBlockhash,
		},
	}
}

func TestDltIngressTxService_Svm_PrepareTransaction_Success(t *testing.T) {
	svc, _, txBuilderRegistry, _, _, blockhashProvider, svmClientRegistry := newService()
	ctx := context.Background()

	blockhashProvider.On("GetRecentBlockhash", ctx, validSvmNetworkId).Return(validSvmBlockhash, nil)

	svmClient := new(svmmocks.SvmClientMock)
	svmClientRegistry.On("GetClientForNetworkId", validSvmNetworkId).Return(svmClient, nil)
	svmClient.On("GetRecentPrioritizationFees", ctx, []string{}).Return([]uint64{1000, 2000, 3000}, nil)

	txResp := validSvmTxResponse()
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.SVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(txResp, nil)
	builder.On("OverrideCuPrice", mock.Anything, mock.Anything, mock.Anything).Return(txResp, nil)

	result, err := svc.PrepareTransaction(ctx, TransactionRequest{
		SenderDltAccountId: validSvmSender,
		SmartContractId:    validSvmContractId,
		SmartContractName:  validSmartContractName,
		MethodName:         validMethodName,
		MethodArgs:         map[string]any{},
		NetworkId:          validSvmNetworkId,
	}, svmNetwork(amount.Zero()))

	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.SVMTransactionResponse)
	assert.Equal(t, txResp, result)
	blockhashProvider.AssertExpectations(t)
	svmClientRegistry.AssertExpectations(t)
	svmClient.AssertExpectations(t)
	builder.AssertExpectations(t)
}

func TestDltIngressTxService_Svm_PrepareTransaction_SeedsCuLimitFromMaxCuLimit(t *testing.T) {
	svc, _, txBuilderRegistry, _, _, blockhashProvider, svmClientRegistry := newService()
	ctx := context.Background()

	// GasLimitMultiplier is zero, so estimation is skipped entirely and the seed passed into
	// BuildTransaction is exactly what reaches the signer — it must come from MaxCuLimit, not
	// from the EVM-sized GasLimit.
	maxCuLimit, _ := amount.NewFromString("1400000")
	network := svmNetwork(amount.Zero())
	network.MaxCuLimit = *maxCuLimit
	network.GasLimit = *amount.Zero()

	blockhashProvider.On("GetRecentBlockhash", ctx, validSvmNetworkId).Return(validSvmBlockhash, nil)
	svmClient := new(svmmocks.SvmClientMock)
	svmClientRegistry.On("GetClientForNetworkId", validSvmNetworkId).Return(svmClient, nil)
	svmClient.On("GetRecentPrioritizationFees", ctx, []string{}).Return([]uint64{1000}, nil)

	txResp := validSvmTxResponse()
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.SVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.MatchedBy(func(req *contracttransactionbuilder.BuildTransactionRequest) bool {
		return req.SVMBuildTransactionRequest != nil &&
			req.SVMBuildTransactionRequest.CuLimit != nil &&
			req.SVMBuildTransactionRequest.CuLimit.Equal(*maxCuLimit)
	})).Return(txResp, nil)
	builder.On("OverrideCuPrice", mock.Anything, mock.Anything, mock.Anything).Return(txResp, nil)

	result, err := svc.PrepareTransaction(ctx, TransactionRequest{
		SenderDltAccountId: validSvmSender,
		SmartContractId:    validSvmContractId,
		SmartContractName:  validSmartContractName,
		MethodName:         validMethodName,
		MethodArgs:         map[string]any{},
		NetworkId:          validSvmNetworkId,
	}, network)

	assert.Nil(t, err)
	assert.Equal(t, txResp, result)
	builder.AssertExpectations(t)
}

func TestDltIngressTxService_Svm_PrepareTransaction_BlockhashError_ReturnsError(t *testing.T) {
	svc, _, txBuilderRegistry, _, _, blockhashProvider, _ := newService()
	ctx := context.Background()

	blockhashProvider.On("GetRecentBlockhash", ctx, validSvmNetworkId).Return("", assert.AnError)
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.SVM, validSmartContractName, builder)

	_, err := svc.PrepareTransaction(ctx, TransactionRequest{
		SenderDltAccountId: validSvmSender,
		SmartContractId:    validSvmContractId,
		SmartContractName:  validSmartContractName,
		MethodName:         validMethodName,
		MethodArgs:         map[string]any{},
		NetworkId:          validSvmNetworkId,
	}, svmNetwork(amount.Zero()))

	assert.ErrorIs(t, err, assert.AnError)
	builder.AssertNotCalled(t, "BuildTransaction")
}

func TestDltIngressTxService_Svm_PrepareTransaction_CuPriceClientError_ReturnsError(t *testing.T) {
	svc, _, txBuilderRegistry, _, _, blockhashProvider, svmClientRegistry := newService()
	ctx := context.Background()

	blockhashProvider.On("GetRecentBlockhash", ctx, validSvmNetworkId).Return(validSvmBlockhash, nil)
	svmClientRegistry.On("GetClientForNetworkId", validSvmNetworkId).Return((*svmmocks.SvmClientMock)(nil), assert.AnError)
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.SVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(validSvmTxResponse(), nil)

	_, err := svc.PrepareTransaction(ctx, TransactionRequest{
		SenderDltAccountId: validSvmSender,
		SmartContractId:    validSvmContractId,
		SmartContractName:  validSmartContractName,
		MethodName:         validMethodName,
		MethodArgs:         map[string]any{},
		NetworkId:          validSvmNetworkId,
	}, svmNetwork(amount.Zero()))

	assert.ErrorIs(t, err, assert.AnError)
	builder.AssertNotCalled(t, "OverrideCuPrice")
	builder.AssertExpectations(t)
}

func TestDltIngressTxService_Svm_PrepareTransaction_CuPriceFetchError_ReturnsError(t *testing.T) {
	svc, _, txBuilderRegistry, _, _, blockhashProvider, svmClientRegistry := newService()
	ctx := context.Background()

	blockhashProvider.On("GetRecentBlockhash", ctx, validSvmNetworkId).Return(validSvmBlockhash, nil)
	svmClient := new(svmmocks.SvmClientMock)
	svmClientRegistry.On("GetClientForNetworkId", validSvmNetworkId).Return(svmClient, nil)
	svmClient.On("GetRecentPrioritizationFees", ctx, []string{}).Return(([]uint64)(nil), assert.AnError)
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.SVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(validSvmTxResponse(), nil)

	_, err := svc.PrepareTransaction(ctx, TransactionRequest{
		SenderDltAccountId: validSvmSender,
		SmartContractId:    validSvmContractId,
		SmartContractName:  validSmartContractName,
		MethodName:         validMethodName,
		MethodArgs:         map[string]any{},
		NetworkId:          validSvmNetworkId,
	}, svmNetwork(amount.Zero()))

	assert.ErrorIs(t, err, assert.AnError)
	builder.AssertNotCalled(t, "OverrideCuPrice")
	builder.AssertExpectations(t)
}

func TestDltIngressTxService_Svm_PrepareTransaction_GasEstimation_Success(t *testing.T) {
	svc, _, txBuilderRegistry, _, gasEstimatorRegistry, blockhashProvider, svmClientRegistry := newService()
	ctx := context.Background()
	network := svmNetwork(amount.Ten())

	blockhashProvider.On("GetRecentBlockhash", ctx, validSvmNetworkId).Return(validSvmBlockhash, nil)
	svmClient := new(svmmocks.SvmClientMock)
	svmClientRegistry.On("GetClientForNetworkId", validSvmNetworkId).Return(svmClient, nil)
	svmClient.On("GetRecentPrioritizationFees", ctx, []string{}).Return([]uint64{5000}, nil)

	txResp := validSvmTxResponse()
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.SVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(txResp, nil)
	builder.On("OverrideCuPrice", mock.Anything, mock.Anything, mock.Anything).Return(txResp, nil)

	estimation, _ := amount.NewFromString("200000")
	estimator := new(gasmock.TransactionGasEstimatorMock)
	gasEstimatorRegistry.Register(common.SVM, estimator)
	estimator.On("EstimateGas", ctx, mock.MatchedBy(func(req transactiongasestimator.EstimationRequest) bool {
		return req.NetworkId == validSvmNetworkId
	})).Return(&transactiongasestimator.EstimationResponse{Estimation: estimation}, nil)

	expectedCuLimit := estimation.MulWithDecimals(network.GasLimitMultiplier, 0)
	overrideResp := validSvmTxResponse()
	overrideResp.SVMTransactionResponse.CuLimit = &expectedCuLimit
	builder.On("OverrideGasLimit", mock.Anything, mock.Anything, mock.MatchedBy(func(a *amount.Amount) bool {
		return a.Equal(expectedCuLimit)
	})).Return(overrideResp, nil).Once()

	result, err := svc.PrepareTransaction(ctx, TransactionRequest{
		SenderDltAccountId: validSvmSender,
		SmartContractId:    validSvmContractId,
		SmartContractName:  validSmartContractName,
		MethodName:         validMethodName,
		MethodArgs:         map[string]any{},
		NetworkId:          validSvmNetworkId,
	}, network)

	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, overrideResp, result)
	estimator.AssertExpectations(t)
	builder.AssertExpectations(t)
}

func TestDltIngressTxService_Svm_PrepareTransaction_GasEstimation_BlockhashNotFound_RefetchesAndRetries(t *testing.T) {
	svc, _, txBuilderRegistry, _, gasEstimatorRegistry, blockhashProvider, svmClientRegistry := newService()
	ctx := context.Background()
	network := svmNetwork(amount.Ten())
	const refreshedBlockhash = "ESymwgTNX1j3E4qhKfJAUE41nBWEwXufoYryPbkde5ci"

	// The transaction is first built with the (stale) cached blockhash...
	blockhashProvider.On("GetRecentBlockhash", ctx, validSvmNetworkId).Return(validSvmBlockhash, nil).Once()
	svmClient := new(svmmocks.SvmClientMock)
	svmClientRegistry.On("GetClientForNetworkId", validSvmNetworkId).Return(svmClient, nil)
	svmClient.On("GetRecentPrioritizationFees", ctx, []string{}).Return([]uint64{5000}, nil)

	txResp := validSvmTxResponse()
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.SVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.MatchedBy(func(req *contracttransactionbuilder.BuildTransactionRequest) bool {
		return req.SVMBuildTransactionRequest.RecentBlockHash == validSvmBlockhash
	})).Return(txResp, nil).Once()
	builder.On("OverrideCuPrice", mock.Anything, mock.Anything, mock.Anything).Return(txResp, nil).Once()

	// ...simulation rejects it as stale, so the blockhash is invalidated and refetched...
	refreshedTxResp := &portcommon.TransactionResponse{
		SVMTransactionResponse: &portcommon.SVMTransactionResponse{
			SerializedTransaction: "AQABAgIDBAUGBwgJ",
			FeePayer:              validSvmSender,
			RecentBlockhash:       refreshedBlockhash,
		},
	}
	blockhashProvider.On("InvalidateRecentBlockhash", ctx, validSvmNetworkId).Return(nil).Once()
	blockhashProvider.On("GetRecentBlockhash", ctx, validSvmNetworkId).Return(refreshedBlockhash, nil).Once()
	builder.On("BuildTransaction", mock.MatchedBy(func(req *contracttransactionbuilder.BuildTransactionRequest) bool {
		return req.SVMBuildTransactionRequest.RecentBlockHash == refreshedBlockhash
	})).Return(refreshedTxResp, nil).Once()

	// ...and the retry, simulated against the fresh blockhash, succeeds.
	estimation, _ := amount.NewFromString("200000")
	estimator := new(gasmock.TransactionGasEstimatorMock)
	gasEstimatorRegistry.Register(common.SVM, estimator)
	estimator.On("EstimateGas", ctx, mock.MatchedBy(func(req transactiongasestimator.EstimationRequest) bool {
		return req.Transaction.SVMTransactionResponse.RecentBlockhash == validSvmBlockhash
	})).Return((*transactiongasestimator.EstimationResponse)(nil), fmt.Errorf("simulation failed: %w", svmmocks.ErrBlockhashNotFound)).Once()
	estimator.On("EstimateGas", ctx, mock.MatchedBy(func(req transactiongasestimator.EstimationRequest) bool {
		return req.Transaction.SVMTransactionResponse.RecentBlockhash == refreshedBlockhash
	})).Return(&transactiongasestimator.EstimationResponse{Estimation: estimation}, nil).Once()

	expectedCuLimit := estimation.MulWithDecimals(network.GasLimitMultiplier, 0)
	overrideResp := &portcommon.TransactionResponse{
		SVMTransactionResponse: &portcommon.SVMTransactionResponse{
			SerializedTransaction: refreshedTxResp.SerializedTransaction,
			FeePayer:              refreshedTxResp.FeePayer,
			RecentBlockhash:       refreshedBlockhash,
			CuLimit:               &expectedCuLimit,
		},
	}
	builder.On("OverrideGasLimit", mock.MatchedBy(func(req contracttransactionbuilder.BuildTransactionRequest) bool {
		return req.SVMBuildTransactionRequest.RecentBlockHash == refreshedBlockhash
	}), mock.Anything, mock.MatchedBy(func(a *amount.Amount) bool {
		return a.Equal(expectedCuLimit)
	})).Return(overrideResp, nil).Once()

	result, err := svc.PrepareTransaction(ctx, TransactionRequest{
		SenderDltAccountId: validSvmSender,
		SmartContractId:    validSvmContractId,
		SmartContractName:  validSmartContractName,
		MethodName:         validMethodName,
		MethodArgs:         map[string]any{},
		NetworkId:          validSvmNetworkId,
	}, network)

	assert.Nil(t, err)
	assert.Equal(t, overrideResp, result)
	blockhashProvider.AssertExpectations(t)
	estimator.AssertExpectations(t)
	builder.AssertExpectations(t)
}

func TestDltIngressTxService_Svm_PrepareTransaction_GasEstimation_BlockhashNotFound_RefetchFails_FallsBackToClearedCuLimit(t *testing.T) {
	svc, _, txBuilderRegistry, _, gasEstimatorRegistry, blockhashProvider, svmClientRegistry := newService()
	ctx := context.Background()
	network := svmNetwork(amount.Ten())

	blockhashProvider.On("GetRecentBlockhash", ctx, validSvmNetworkId).Return(validSvmBlockhash, nil).Once()
	blockhashProvider.On("InvalidateRecentBlockhash", ctx, validSvmNetworkId).Return(nil).Once()
	blockhashProvider.On("GetRecentBlockhash", ctx, validSvmNetworkId).Return("", assert.AnError).Once()

	svmClient := new(svmmocks.SvmClientMock)
	svmClientRegistry.On("GetClientForNetworkId", validSvmNetworkId).Return(svmClient, nil)
	svmClient.On("GetRecentPrioritizationFees", ctx, []string{}).Return([]uint64{5000}, nil)

	txResp := validSvmTxResponse()
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.SVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(txResp, nil).Once()
	builder.On("OverrideCuPrice", mock.Anything, mock.Anything, mock.Anything).Return(txResp, nil).Once()

	estimator := new(gasmock.TransactionGasEstimatorMock)
	gasEstimatorRegistry.Register(common.SVM, estimator)
	estimator.On("EstimateGas", ctx, mock.Anything).
		Return((*transactiongasestimator.EstimationResponse)(nil), fmt.Errorf("simulation failed: %w", svmmocks.ErrBlockhashNotFound)).Once()

	clearedResp := validSvmTxResponse()
	clearedResp.SVMTransactionResponse.CuLimit = amount.Zero()
	builder.On("OverrideGasLimit", mock.Anything, mock.Anything, mock.MatchedBy(func(a *amount.Amount) bool {
		return a.IsZero()
	})).Return(clearedResp, nil).Once()

	result, err := svc.PrepareTransaction(ctx, TransactionRequest{
		SenderDltAccountId: validSvmSender,
		SmartContractId:    validSvmContractId,
		SmartContractName:  validSmartContractName,
		MethodName:         validMethodName,
		MethodArgs:         map[string]any{},
		NetworkId:          validSvmNetworkId,
	}, network)

	assert.Nil(t, err)
	assert.Equal(t, clearedResp, result)
	blockhashProvider.AssertExpectations(t)
	estimator.AssertExpectations(t)
	builder.AssertExpectations(t)
}

func TestDltIngressTxService_Svm_PrepareTransaction_GasEstimation_Skipped_WhenMultiplierBelowOne(t *testing.T) {
	svc, _, txBuilderRegistry, _, _, blockhashProvider, svmClientRegistry := newService()
	ctx := context.Background()

	blockhashProvider.On("GetRecentBlockhash", ctx, validSvmNetworkId).Return(validSvmBlockhash, nil)
	svmClient := new(svmmocks.SvmClientMock)
	svmClientRegistry.On("GetClientForNetworkId", validSvmNetworkId).Return(svmClient, nil)
	svmClient.On("GetRecentPrioritizationFees", ctx, []string{}).Return([]uint64{1000}, nil)

	txResp := validSvmTxResponse()
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.SVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(txResp, nil)
	builder.On("OverrideCuPrice", mock.Anything, mock.Anything, mock.Anything).Return(txResp, nil)

	result, err := svc.PrepareTransaction(ctx, TransactionRequest{
		SenderDltAccountId: validSvmSender,
		SmartContractId:    validSvmContractId,
		SmartContractName:  validSmartContractName,
		MethodName:         validMethodName,
		MethodArgs:         map[string]any{},
		NetworkId:          validSvmNetworkId,
	}, svmNetwork(amount.Zero()))

	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, txResp, result)
	builder.AssertNotCalled(t, "OverrideGasLimit")
	builder.AssertExpectations(t)
}

func TestDltIngressTxService_Svm_PrepareTransaction_CuPriceIsCappedAtMaxCuPrice(t *testing.T) {
	svc, _, txBuilderRegistry, _, _, blockhashProvider, svmClientRegistry := newService()
	ctx := context.Background()

	maxCuPrice, _ := amount.NewFromString("500")
	network := svmNetwork(amount.Zero())
	network.MaxCuPrice = *maxCuPrice

	blockhashProvider.On("GetRecentBlockhash", ctx, validSvmNetworkId).Return(validSvmBlockhash, nil)
	svmClient := new(svmmocks.SvmClientMock)
	svmClientRegistry.On("GetClientForNetworkId", validSvmNetworkId).Return(svmClient, nil)
	// p75 of [1000, 2000, 3000] = 3000, which exceeds maxCuPrice of 500
	svmClient.On("GetRecentPrioritizationFees", ctx, []string{}).Return([]uint64{1000, 2000, 3000}, nil)

	txResp := validSvmTxResponse()
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.SVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(txResp, nil)
	builder.On("OverrideCuPrice", mock.MatchedBy(func(req contracttransactionbuilder.BuildTransactionRequest) bool {
		return req.CuPrice != nil && req.CuPrice.Equal(*maxCuPrice)
	}), mock.Anything, mock.MatchedBy(func(a *amount.Amount) bool {
		return a.Equal(*maxCuPrice)
	})).Return(txResp, nil)

	result, err := svc.PrepareTransaction(ctx, TransactionRequest{
		SenderDltAccountId: validSvmSender,
		SmartContractId:    validSvmContractId,
		SmartContractName:  validSmartContractName,
		MethodName:         validMethodName,
		MethodArgs:         map[string]any{},
		NetworkId:          validSvmNetworkId,
	}, network)

	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, txResp, result)
	svmClient.AssertExpectations(t)
	builder.AssertExpectations(t)
}

func TestDltIngressTxService_Svm_PrepareTransaction_MaxCuLimitUnset_NoCapApplied(t *testing.T) {
	svc, _, txBuilderRegistry, _, gasEstimatorRegistry, blockhashProvider, svmClientRegistry := newService()
	ctx := context.Background()

	// MaxCuLimit left unset means "no cap", never "cap at zero".
	network := svmNetwork(amount.Ten())

	blockhashProvider.On("GetRecentBlockhash", ctx, validSvmNetworkId).Return(validSvmBlockhash, nil)
	svmClient := new(svmmocks.SvmClientMock)
	svmClientRegistry.On("GetClientForNetworkId", validSvmNetworkId).Return(svmClient, nil)
	svmClient.On("GetRecentPrioritizationFees", ctx, []string{}).Return([]uint64{1000}, nil)

	txResp := validSvmTxResponse()
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.SVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(txResp, nil)
	builder.On("OverrideCuPrice", mock.Anything, mock.Anything, mock.Anything).Return(txResp, nil)

	// 2,000,000 x 10 = 20,000,000, far above Solana's real 1.4M ceiling, but no cap is configured.
	estimation, _ := amount.NewFromString("2000000")
	estimator := new(gasmock.TransactionGasEstimatorMock)
	gasEstimatorRegistry.Register(common.SVM, estimator)
	estimator.On("EstimateGas", ctx, mock.Anything).
		Return(&transactiongasestimator.EstimationResponse{Estimation: estimation}, nil)

	expectedCuLimit := estimation.MulWithDecimals(network.GasLimitMultiplier, 0)
	overrideResp := validSvmTxResponse()
	overrideResp.SVMTransactionResponse.CuLimit = &expectedCuLimit
	builder.On("OverrideGasLimit", mock.Anything, mock.Anything, mock.MatchedBy(func(a *amount.Amount) bool {
		return a.Equal(expectedCuLimit)
	})).Return(overrideResp, nil).Once()

	result, err := svc.PrepareTransaction(ctx, TransactionRequest{
		SenderDltAccountId: validSvmSender,
		SmartContractId:    validSvmContractId,
		SmartContractName:  validSmartContractName,
		MethodName:         validMethodName,
		MethodArgs:         map[string]any{},
		NetworkId:          validSvmNetworkId,
	}, network)

	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint64(20000000), result.SVMTransactionResponse.CuLimit.RawValue().Uint64())
	estimator.AssertExpectations(t)
	builder.AssertExpectations(t)
}

func TestDltIngressTxService_Svm_PrepareTransaction_CuLimitBelowMax_PassesThrough(t *testing.T) {
	svc, _, txBuilderRegistry, _, gasEstimatorRegistry, blockhashProvider, svmClientRegistry := newService()
	ctx := context.Background()

	network := svmNetwork(amount.One())
	maxCuLimit, _ := amount.NewFromString("1400000")
	network.MaxCuLimit = *maxCuLimit

	blockhashProvider.On("GetRecentBlockhash", ctx, validSvmNetworkId).Return(validSvmBlockhash, nil)
	svmClient := new(svmmocks.SvmClientMock)
	svmClientRegistry.On("GetClientForNetworkId", validSvmNetworkId).Return(svmClient, nil)
	svmClient.On("GetRecentPrioritizationFees", ctx, []string{}).Return([]uint64{1000}, nil)

	txResp := validSvmTxResponse()
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.SVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(txResp, nil)
	builder.On("OverrideCuPrice", mock.Anything, mock.Anything, mock.Anything).Return(txResp, nil)

	// 1,350,000 sits just below the 1,400,000 ceiling, so it must be applied untouched.
	estimation, _ := amount.NewFromString("1350000")
	estimator := new(gasmock.TransactionGasEstimatorMock)
	gasEstimatorRegistry.Register(common.SVM, estimator)
	estimator.On("EstimateGas", ctx, mock.Anything).
		Return(&transactiongasestimator.EstimationResponse{Estimation: estimation}, nil)

	overrideResp := validSvmTxResponse()
	overrideResp.SVMTransactionResponse.CuLimit = estimation
	builder.On("OverrideGasLimit", mock.Anything, mock.Anything, mock.MatchedBy(func(a *amount.Amount) bool {
		return a.Equal(*estimation)
	})).Return(overrideResp, nil).Once()

	result, err := svc.PrepareTransaction(ctx, TransactionRequest{
		SenderDltAccountId: validSvmSender,
		SmartContractId:    validSvmContractId,
		SmartContractName:  validSmartContractName,
		MethodName:         validMethodName,
		MethodArgs:         map[string]any{},
		NetworkId:          validSvmNetworkId,
	}, network)

	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint64(1350000), result.SVMTransactionResponse.CuLimit.RawValue().Uint64())
	estimator.AssertExpectations(t)
	builder.AssertExpectations(t)
}

func TestDltIngressTxService_Svm_PrepareTransaction_CuLimitAboveMax_FailsFast(t *testing.T) {
	svc, _, txBuilderRegistry, _, gasEstimatorRegistry, blockhashProvider, svmClientRegistry := newService()
	ctx := context.Background()

	network := svmNetwork(amount.One())
	maxCuLimit, _ := amount.NewFromString("1400000")
	network.MaxCuLimit = *maxCuLimit

	blockhashProvider.On("GetRecentBlockhash", ctx, validSvmNetworkId).Return(validSvmBlockhash, nil)
	svmClient := new(svmmocks.SvmClientMock)
	svmClientRegistry.On("GetClientForNetworkId", validSvmNetworkId).Return(svmClient, nil)
	svmClient.On("GetRecentPrioritizationFees", ctx, []string{}).Return([]uint64{1000}, nil)

	txResp := validSvmTxResponse()
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.SVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(txResp, nil)
	builder.On("OverrideCuPrice", mock.Anything, mock.Anything, mock.Anything).Return(txResp, nil)

	// 2,000,000 exceeds the 1,400,000 ceiling: the transaction must be refused, not clamped.
	estimation, _ := amount.NewFromString("2000000")
	estimator := new(gasmock.TransactionGasEstimatorMock)
	gasEstimatorRegistry.Register(common.SVM, estimator)
	estimator.On("EstimateGas", ctx, mock.Anything).
		Return(&transactiongasestimator.EstimationResponse{Estimation: estimation}, nil)

	result, err := svc.PrepareTransaction(ctx, TransactionRequest{
		SenderDltAccountId: validSvmSender,
		SmartContractId:    validSvmContractId,
		SmartContractName:  validSmartContractName,
		MethodName:         validMethodName,
		MethodArgs:         map[string]any{},
		NetworkId:          validSvmNetworkId,
	}, network)

	// The error must propagate rather than be degraded to a warning: swallowing it here would
	// let the un-overridden seed reach the signer.
	assert.Nil(t, result)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "2000000")
	assert.Contains(t, err.Error(), "1400000")
	builder.AssertNotCalled(t, "OverrideGasLimit")
	estimator.AssertExpectations(t)
}

func TestDltIngressTxService_Svm_PrepareTransaction_MultipliedCuLimitAboveMax_SkipsOverride(t *testing.T) {
	svc, _, txBuilderRegistry, _, gasEstimatorRegistry, blockhashProvider, svmClientRegistry := newService()
	ctx := context.Background()

	multiplier, _ := amount.NewFromString("2")
	network := svmNetwork(multiplier)
	maxCuLimit, _ := amount.NewFromString("1400000")
	network.MaxCuLimit = *maxCuLimit

	blockhashProvider.On("GetRecentBlockhash", ctx, validSvmNetworkId).Return(validSvmBlockhash, nil)
	svmClient := new(svmmocks.SvmClientMock)
	svmClientRegistry.On("GetClientForNetworkId", validSvmNetworkId).Return(svmClient, nil)
	svmClient.On("GetRecentPrioritizationFees", ctx, []string{}).Return([]uint64{1000}, nil)

	txResp := validSvmTxResponse()
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.SVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(txResp, nil)
	builder.On("OverrideCuPrice", mock.Anything, mock.Anything, mock.Anything).Return(txResp, nil)

	// 1,000,000 sits below the 1,400,000 ceiling on its own, but doubled by the multiplier it
	// would land at 2,000,000, above the ceiling.
	estimation, _ := amount.NewFromString("1000000")
	estimator := new(gasmock.TransactionGasEstimatorMock)
	gasEstimatorRegistry.Register(common.SVM, estimator)
	estimator.On("EstimateGas", ctx, mock.Anything).
		Return(&transactiongasestimator.EstimationResponse{Estimation: estimation}, nil)

	result, err := svc.PrepareTransaction(ctx, TransactionRequest{
		SenderDltAccountId: validSvmSender,
		SmartContractId:    validSvmContractId,
		SmartContractName:  validSmartContractName,
		MethodName:         validMethodName,
		MethodArgs:         map[string]any{},
		NetworkId:          validSvmNetworkId,
	}, network)

	// The raw estimate is within the cap, so this must not be treated as a failure: the
	// override is simply skipped and the original, un-overridden transaction goes through.
	assert.Nil(t, err)
	assert.Equal(t, txResp, result)
	builder.AssertNotCalled(t, "OverrideGasLimit")
	estimator.AssertExpectations(t)
	builder.AssertExpectations(t)
}

func TestDltIngressTxService_Svm_PrepareTransaction_GasEstimation_Error_OmitsCuLimit(t *testing.T) {
	svc, _, txBuilderRegistry, _, gasEstimatorRegistry, blockhashProvider, svmClientRegistry := newService()
	ctx := context.Background()
	network := svmNetwork(amount.One())

	blockhashProvider.On("GetRecentBlockhash", ctx, validSvmNetworkId).Return(validSvmBlockhash, nil)
	svmClient := new(svmmocks.SvmClientMock)
	svmClientRegistry.On("GetClientForNetworkId", validSvmNetworkId).Return(svmClient, nil)
	svmClient.On("GetRecentPrioritizationFees", ctx, []string{}).Return([]uint64{1000}, nil)

	txResp := validSvmTxResponse()
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.SVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(txResp, nil)
	builder.On("OverrideCuPrice", mock.Anything, mock.Anything, mock.Anything).Return(txResp, nil)

	estimator := new(gasmock.TransactionGasEstimatorMock)
	gasEstimatorRegistry.Register(common.SVM, estimator)
	estimator.On("EstimateGas", ctx, mock.Anything).
		Return((*transactiongasestimator.EstimationResponse)(nil), fmt.Errorf("simulation failed"))

	// A failed simulation must not let the oversized pre-simulation seed reach the signer: the
	// limit is cleared to zero (not nil — CuLimit is persisted to a NOT NULL column), which the
	// builder treats the same as an absent limit.
	omittedResp := validSvmTxResponse()
	builder.On("OverrideGasLimit", mock.Anything, mock.Anything, mock.MatchedBy(func(a *amount.Amount) bool {
		return a != nil && a.IsZero()
	})).Return(omittedResp, nil).Once()

	result, err := svc.PrepareTransaction(ctx, TransactionRequest{
		SenderDltAccountId: validSvmSender,
		SmartContractId:    validSvmContractId,
		SmartContractName:  validSmartContractName,
		MethodName:         validMethodName,
		MethodArgs:         map[string]any{},
		NetworkId:          validSvmNetworkId,
	}, network)

	assert.Nil(t, err)
	assert.Equal(t, omittedResp, result)
	estimator.AssertExpectations(t)
	builder.AssertExpectations(t)
}

// svmTxResponseWithAccounts lays out one account of each kind in Solana's compiled-message order:
// signed writable (fee payer), signed readonly, unsigned writable, unsigned readonly (program).
func svmTxResponseWithAccounts() (*portcommon.TransactionResponse, []string) {
	const (
		signedReadonly   = "4Nd1mBQtrMJVYVfKf2PJy9NZUZdTAsp7D4xWLs4gDB4T"
		unsignedWritable = "So11111111111111111111111111111111111111112"
	)
	txResp := validSvmTxResponse()
	txResp.SVMTransactionResponse.Header = portcommon.SVMMessageHeader{
		NumRequiredSignatures:       2,
		NumReadonlySignedAccounts:   1,
		NumReadonlyUnsignedAccounts: 1,
	}
	txResp.SVMTransactionResponse.AccountKeys = []string{validSvmSender, signedReadonly, unsignedWritable, validSvmContractId}
	return txResp, []string{validSvmSender, unsignedWritable}
}

func TestDltIngressTxService_Svm_PrepareTransaction_FetchesFeesForWritableAccounts(t *testing.T) {
	svc, _, txBuilderRegistry, _, _, blockhashProvider, svmClientRegistry := newService()
	ctx := context.Background()

	blockhashProvider.On("GetRecentBlockhash", ctx, validSvmNetworkId).Return(validSvmBlockhash, nil)
	svmClient := new(svmmocks.SvmClientMock)
	svmClientRegistry.On("GetClientForNetworkId", validSvmNetworkId).Return(svmClient, nil)

	txResp, writableAccounts := svmTxResponseWithAccounts()
	// The fee estimate must reflect contention on the accounts this transaction write-locks,
	// not overall ledger activity (which is what an empty account list asks the node for).
	svmClient.On("GetRecentPrioritizationFees", ctx, writableAccounts).Return([]uint64{1000, 2000, 3000}, nil).Once()

	// p75 of [1000, 2000, 3000] = 3000
	expectedCuPrice, _ := amount.NewFromString("3000")
	pricedResp := validSvmTxResponse()
	pricedResp.SVMTransactionResponse.CuPrice = expectedCuPrice

	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.SVM, validSmartContractName, builder)
	// Accounts are only resolved by the build itself, so the first build cannot carry a price yet.
	builder.On("BuildTransaction", mock.MatchedBy(func(req *contracttransactionbuilder.BuildTransactionRequest) bool {
		return req.SVMBuildTransactionRequest != nil && req.SVMBuildTransactionRequest.CuPrice == nil
	})).Return(txResp, nil).Once()
	builder.On("OverrideCuPrice", mock.MatchedBy(func(req contracttransactionbuilder.BuildTransactionRequest) bool {
		return req.CuPrice != nil && req.CuPrice.Equal(*expectedCuPrice)
	}), *txResp, mock.MatchedBy(func(a *amount.Amount) bool {
		return a.Equal(*expectedCuPrice)
	})).Return(pricedResp, nil).Once()

	result, err := svc.PrepareTransaction(ctx, TransactionRequest{
		SenderDltAccountId: validSvmSender,
		SmartContractId:    validSvmContractId,
		SmartContractName:  validSmartContractName,
		MethodName:         validMethodName,
		MethodArgs:         map[string]any{},
		NetworkId:          validSvmNetworkId,
	}, svmNetwork(amount.Zero()))

	assert.Nil(t, err)
	assert.Equal(t, pricedResp, result)
	svmClient.AssertExpectations(t)
	builder.AssertExpectations(t)
}

func TestDltIngressTxService_Svm_PrepareTransaction_CuPriceSurvivesGasLimitOverride(t *testing.T) {
	svc, _, txBuilderRegistry, _, gasEstimatorRegistry, blockhashProvider, svmClientRegistry := newService()
	ctx := context.Background()
	network := svmNetwork(amount.Ten())

	blockhashProvider.On("GetRecentBlockhash", ctx, validSvmNetworkId).Return(validSvmBlockhash, nil)
	svmClient := new(svmmocks.SvmClientMock)
	svmClientRegistry.On("GetClientForNetworkId", validSvmNetworkId).Return(svmClient, nil)
	txResp, writableAccounts := svmTxResponseWithAccounts()
	svmClient.On("GetRecentPrioritizationFees", ctx, writableAccounts).Return([]uint64{5000}, nil)

	expectedCuPrice, _ := amount.NewFromString("5000")
	pricedResp := validSvmTxResponse()
	pricedResp.SVMTransactionResponse.CuPrice = expectedCuPrice

	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.SVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(txResp, nil).Once()
	builder.On("OverrideCuPrice", mock.Anything, *txResp, mock.Anything).Return(pricedResp, nil).Once()

	estimation, _ := amount.NewFromString("200000")
	estimator := new(gasmock.TransactionGasEstimatorMock)
	gasEstimatorRegistry.Register(common.SVM, estimator)
	// The simulation must run on the priced transaction, so it consumes what will actually be submitted.
	estimator.On("EstimateGas", ctx, mock.MatchedBy(func(req transactiongasestimator.EstimationRequest) bool {
		return req.Transaction == pricedResp
	})).Return(&transactiongasestimator.EstimationResponse{Estimation: estimation}, nil)

	// OverrideGasLimit rebuilds from the request, so the price must be on the request it receives
	// or the final transaction would silently drop it.
	expectedCuLimit := estimation.MulWithDecimals(network.GasLimitMultiplier, 0)
	finalResp := validSvmTxResponse()
	finalResp.SVMTransactionResponse.CuPrice = expectedCuPrice
	finalResp.SVMTransactionResponse.CuLimit = &expectedCuLimit
	builder.On("OverrideGasLimit", mock.MatchedBy(func(req contracttransactionbuilder.BuildTransactionRequest) bool {
		return req.CuPrice != nil && req.CuPrice.Equal(*expectedCuPrice)
	}), *pricedResp, mock.MatchedBy(func(a *amount.Amount) bool {
		return a.Equal(expectedCuLimit)
	})).Return(finalResp, nil).Once()

	result, err := svc.PrepareTransaction(ctx, TransactionRequest{
		SenderDltAccountId: validSvmSender,
		SmartContractId:    validSvmContractId,
		SmartContractName:  validSmartContractName,
		MethodName:         validMethodName,
		MethodArgs:         map[string]any{},
		NetworkId:          validSvmNetworkId,
	}, network)

	assert.Nil(t, err)
	assert.Equal(t, finalResp, result)
	estimator.AssertExpectations(t)
	builder.AssertExpectations(t)
}

func TestDltIngressTxService_Svm_PrepareTransaction_OverrideCuPriceError_ReturnsError(t *testing.T) {
	svc, _, txBuilderRegistry, _, gasEstimatorRegistry, blockhashProvider, svmClientRegistry := newService()
	ctx := context.Background()

	blockhashProvider.On("GetRecentBlockhash", ctx, validSvmNetworkId).Return(validSvmBlockhash, nil)
	svmClient := new(svmmocks.SvmClientMock)
	svmClientRegistry.On("GetClientForNetworkId", validSvmNetworkId).Return(svmClient, nil)
	svmClient.On("GetRecentPrioritizationFees", ctx, []string{}).Return([]uint64{1000}, nil)

	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.SVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(validSvmTxResponse(), nil)
	builder.On("OverrideCuPrice", mock.Anything, mock.Anything, mock.Anything).Return(nil, assert.AnError)
	estimator := new(gasmock.TransactionGasEstimatorMock)
	gasEstimatorRegistry.Register(common.SVM, estimator)

	result, err := svc.PrepareTransaction(ctx, TransactionRequest{
		SenderDltAccountId: validSvmSender,
		SmartContractId:    validSvmContractId,
		SmartContractName:  validSmartContractName,
		MethodName:         validMethodName,
		MethodArgs:         map[string]any{},
		NetworkId:          validSvmNetworkId,
	}, svmNetwork(amount.Ten()))

	assert.Nil(t, result)
	assert.ErrorIs(t, err, assert.AnError)
	estimator.AssertNotCalled(t, "EstimateGas")
}
