package txservice

import (
	"context"
	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/app/service/txservice"
	"dlt-ingress/src/main/dltingress/domain/common"
	"dlt-ingress/src/main/dltingress/port/contracttransactionbuilder"
	"dlt-ingress/src/main/dltingress/port/nonceprovider"
	"dlt-ingress/src/main/dltingress/port/portcommon"
	"dlt-ingress/src/main/dltingress/port/transactiongasestimator"
	ctbmock "dlt-ingress/src/test/dltingress/port/contracttransactionbuilder"
	ethmocks "dlt-ingress/src/test/dltingress/port/evm"
	npmock "dlt-ingress/src/test/dltingress/port/nonceprovider"
	svmmocks "dlt-ingress/src/test/dltingress/port/svm"
	gasmock "dlt-ingress/src/test/dltingress/port/transanctiongasestimation"
	"fmt"
	"testing"

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

	config.AppConfig = &config.Config{}
	config.AppConfig.DltIngress.Networks = []*config.NetworkConfig{
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
	*txservice.AppService,
	*npmock.NonceProviderMock,
	*contracttransactionbuilder.Registry,
	*ethmocks.EvmClientRegistryMock,
	*transactiongasestimator.Registry,
	*svmmocks.BlockhashProviderMock,
	*svmmocks.SvmClientRegistryMock,
) {
	nonceProvider := new(npmock.NonceProviderMock)
	blockhashProvider := new(svmmocks.BlockhashProviderMock)
	registry := contracttransactionbuilder.NewRegistry()
	ethClientRegistry := new(ethmocks.EvmClientRegistryMock)
	svmClientRegistry := new(svmmocks.SvmClientRegistryMock)
	gasEstimatorRegistry := transactiongasestimator.NewRegistry()
	svc := txservice.NewAppService(nonceProvider, blockhashProvider, *registry, ethClientRegistry, svmClientRegistry, *gasEstimatorRegistry)
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

func validRequest() txservice.TransactionRequest {
	return txservice.TransactionRequest{
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
	network := *config.AppConfig.DltIngress.Networks[0]

	ctx := context.Background()
	nonce, _ := amount.NewFromString("0")
	nonceProvider.On("GetNonce", ctx, &nonceprovider.GetNonceRequest{
		NetworkId:    validNetworkId,
		DltAccountId: validSenderDltAccountId,
	}).Return(nonce, nil)

	networkGasPrice, _ := amount.NewFromString(validGasPrice)
	ethClient := new(ethmocks.EvmClientMock)
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
	network := *config.AppConfig.DltIngress.Networks[0]

	ctx := context.Background()
	nonce, _ := amount.NewFromString("0")
	nonceProvider.On("GetNonce", ctx, mock.Anything).Return(nonce, nil)

	ethClientRegistry.On("GetClientForNetworkId", ctx, validNetworkId).
		Return((*ethmocks.EvmClientMock)(nil), assert.AnError)

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
	network := *config.AppConfig.DltIngress.Networks[0]

	ctx := context.Background()
	nonce, _ := amount.NewFromString("0")
	nonceProvider.On("GetNonce", ctx, mock.Anything).Return(nonce, nil)

	ethClient := new(ethmocks.EvmClientMock)
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
	ethClient := new(ethmocks.EvmClientMock)
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
	network := *config.AppConfig.DltIngress.Networks[0]

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
	network := *config.AppConfig.DltIngress.Networks[0]

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
	network := *config.AppConfig.DltIngress.Networks[0]

	ctx := context.Background()
	nonce, _ := amount.NewFromString("0")
	nonceProvider.On("GetNonce", ctx, mock.Anything).Return(nonce, nil)

	networkGasPrice, _ := amount.NewFromString(validGasPrice)
	ethClient := new(ethmocks.EvmClientMock)
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
		Return((*ethmocks.EvmClientMock)(nil), assert.AnError)

	result, err := svc.PrepareTransaction(ctx, validRequest(), network)

	assert.Nil(t, result)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	ethClientRegistry.AssertExpectations(t)
	builder.AssertNotCalled(t, "BuildTransaction", mock.Anything)
}

func TestDltIngressTxService_Evm_PrepareTransaction_GasEstimation_Success(t *testing.T) {
	svc, nonceProvider, txBuilderRegistry, ethClientRegistry, gasEstimatorRegistry, _, _ := newService()
	network := *config.AppConfig.DltIngress.Networks[0]
	network.GasLimitMultiplier = *amount.One()

	ctx := context.Background()
	nonce, _ := amount.NewFromString("0")
	nonceProvider.On("GetNonce", ctx, mock.Anything).Return(nonce, nil)

	networkGasPrice, _ := amount.NewFromString(validGasPrice)
	ethClient := new(ethmocks.EvmClientMock)
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
	network := *config.AppConfig.DltIngress.Networks[0]
	network.GasLimitMultiplier = *amount.One()

	ctx := context.Background()
	nonce, _ := amount.NewFromString("0")
	nonceProvider.On("GetNonce", ctx, mock.Anything).Return(nonce, nil)

	networkGasPrice, _ := amount.NewFromString(validGasPrice)
	ethClient := new(ethmocks.EvmClientMock)
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
	network := *config.AppConfig.DltIngress.Networks[0]
	network.GasLimitMultiplier = *amount.Zero()

	ctx := context.Background()
	nonce, _ := amount.NewFromString("0")
	nonceProvider.On("GetNonce", ctx, mock.Anything).Return(nonce, nil)

	networkGasPrice, _ := amount.NewFromString(validGasPrice)
	ethClient := new(ethmocks.EvmClientMock)
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
	svmClient.On("GetRecentPrioritizationFees", ctx).Return([]uint64{1000, 2000, 3000}, nil)

	txResp := validSvmTxResponse()
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.SVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(txResp, nil)

	result, err := svc.PrepareTransaction(ctx, txservice.TransactionRequest{
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

func TestDltIngressTxService_Svm_PrepareTransaction_BlockhashError_ReturnsError(t *testing.T) {
	svc, _, txBuilderRegistry, _, _, blockhashProvider, _ := newService()
	ctx := context.Background()

	blockhashProvider.On("GetRecentBlockhash", ctx, validSvmNetworkId).Return("", assert.AnError)
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.SVM, validSmartContractName, builder)

	_, err := svc.PrepareTransaction(ctx, txservice.TransactionRequest{
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

	_, err := svc.PrepareTransaction(ctx, txservice.TransactionRequest{
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

func TestDltIngressTxService_Svm_PrepareTransaction_CuPriceFetchError_ReturnsError(t *testing.T) {
	svc, _, txBuilderRegistry, _, _, blockhashProvider, svmClientRegistry := newService()
	ctx := context.Background()

	blockhashProvider.On("GetRecentBlockhash", ctx, validSvmNetworkId).Return(validSvmBlockhash, nil)
	svmClient := new(svmmocks.SvmClientMock)
	svmClientRegistry.On("GetClientForNetworkId", validSvmNetworkId).Return(svmClient, nil)
	svmClient.On("GetRecentPrioritizationFees", ctx).Return(([]uint64)(nil), assert.AnError)
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.SVM, validSmartContractName, builder)

	_, err := svc.PrepareTransaction(ctx, txservice.TransactionRequest{
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

func TestDltIngressTxService_Svm_PrepareTransaction_GasEstimation_Success(t *testing.T) {
	svc, _, txBuilderRegistry, _, gasEstimatorRegistry, blockhashProvider, svmClientRegistry := newService()
	ctx := context.Background()
	network := svmNetwork(amount.Ten())

	blockhashProvider.On("GetRecentBlockhash", ctx, validSvmNetworkId).Return(validSvmBlockhash, nil)
	svmClient := new(svmmocks.SvmClientMock)
	svmClientRegistry.On("GetClientForNetworkId", validSvmNetworkId).Return(svmClient, nil)
	svmClient.On("GetRecentPrioritizationFees", ctx).Return([]uint64{5000}, nil)

	txResp := validSvmTxResponse()
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.SVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(txResp, nil)

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

	result, err := svc.PrepareTransaction(ctx, txservice.TransactionRequest{
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

func TestDltIngressTxService_Svm_PrepareTransaction_GasEstimation_Skipped_WhenMultiplierBelowOne(t *testing.T) {
	svc, _, txBuilderRegistry, _, _, blockhashProvider, svmClientRegistry := newService()
	ctx := context.Background()

	blockhashProvider.On("GetRecentBlockhash", ctx, validSvmNetworkId).Return(validSvmBlockhash, nil)
	svmClient := new(svmmocks.SvmClientMock)
	svmClientRegistry.On("GetClientForNetworkId", validSvmNetworkId).Return(svmClient, nil)
	svmClient.On("GetRecentPrioritizationFees", ctx).Return([]uint64{1000}, nil)

	txResp := validSvmTxResponse()
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.SVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(txResp, nil)

	result, err := svc.PrepareTransaction(ctx, txservice.TransactionRequest{
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
	network.MaxCuPrice.RawValue().Uint64()

	blockhashProvider.On("GetRecentBlockhash", ctx, validSvmNetworkId).Return(validSvmBlockhash, nil)
	svmClient := new(svmmocks.SvmClientMock)
	svmClientRegistry.On("GetClientForNetworkId", validSvmNetworkId).Return(svmClient, nil)
	// p75 of [1000, 2000, 3000] = 3000, which exceeds maxCuPrice of 500
	svmClient.On("GetRecentPrioritizationFees", ctx).Return([]uint64{1000, 2000, 3000}, nil)

	txResp := validSvmTxResponse()
	builder := new(ctbmock.ContractTransactionBuilderMock)
	txBuilderRegistry.Register(common.SVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.MatchedBy(func(req *contracttransactionbuilder.BuildTransactionRequest) bool {
		return req.CuPrice != nil && req.CuPrice.Equal(*maxCuPrice)
	})).Return(txResp, nil)

	result, err := svc.PrepareTransaction(ctx, txservice.TransactionRequest{
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
