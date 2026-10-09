package signandsend

import (
	"context"
	"testing"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/app/service/txservice"
	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"dlt-ingress/src/main/dltingress/internal/domain/custodykey"
	servicemock "dlt-ingress/src/main/dltingress/internal/domain/custodykey/service/mock"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction"
	bbqmock "dlt-ingress/src/main/dltingress/internal/infra/boundedblockingqueue/mock"
	"dlt-ingress/src/main/dltingress/internal/infra/contracttransactionbuilder"
	ctbmock "dlt-ingress/src/main/dltingress/internal/infra/contracttransactionbuilder/mock"
	"dlt-ingress/src/main/dltingress/internal/infra/custody"
	custodymock "dlt-ingress/src/main/dltingress/internal/infra/custody/mock"
	mock2 "dlt-ingress/src/main/dltingress/internal/infra/evm/mock"
	"dlt-ingress/src/main/dltingress/internal/infra/nonceprovider"
	npmock "dlt-ingress/src/main/dltingress/internal/infra/nonceprovider/mock"
	"dlt-ingress/src/main/dltingress/internal/infra/portcommon"
	"dlt-ingress/src/main/dltingress/internal/infra/repository/transaction/evmtransaction/mock"
	"dlt-ingress/src/main/dltingress/internal/infra/repository/transaction/svmtransaction/mock"
	svmmocks "dlt-ingress/src/main/dltingress/internal/infra/svm"
	"dlt-ingress/src/main/dltingress/internal/infra/transactiongasestimator"
	"dlt-ingress/src/main/dltingress/internal/infra/transanctionsender"
	tsmock "dlt-ingress/src/main/dltingress/internal/infra/transanctionsender/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	testbus "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/test/core/event"
)

const (
	validNetworkId            = "mainnet"
	validDlt                  = "EVM"
	validUrl                  = "http://localhost:8545"
	validChainId              = "1"
	validGasLimit             = "21000"
	validMaxPriorityFeePerGas = "1000000000"
	validMaxFeePerGas         = "2000000000"
	validSenderDltAccountId   = "0x123"
	validSmartContractId      = "0x456"
	validSmartContractName    = "ERC20"
	validMethodName           = "transfer"
	validTxId                 = "0x789"
	validExternalId           = "ext-123"
	validTransactionType      = 0
	validGasPrice             = "1000000000"
	validFeeMultiplier        = "1.2"
	validMaxTxPoolSize        = 100

	validSvmNetworkId    = "solana-mainnet"
	validSvmDlt          = "SVM"
	validSvmUrl          = "https://api.mainnet-beta.solana.com"
	validSvmBlockhash    = "GHtXQBsoZHVnNFa9YevAzFr17DJjgHXk3ycTKD5xD3Zi"
	validSvmSender       = "9xQeWvG816bUx9EPjHmaT23yvVM2ZWbrrpZb9PusVFin"
	validSvmContractId   = "11111111111111111111111111111112"
	validSvmTxId         = "5VfYmGBL5dV1aH6dvB1Yd9k3oQ8Wd2nQ2Yk9eQ7Wd2nQ2Yk9eQ7Wd2nQ2Yk9eQ7Wd2"
	validSvmSerializedTx = "AQABAgIDBAUGBwgJ"
)

func init() {
	typedChainId, _ := amount.NewFromString(validChainId)
	typedGasLimit, _ := amount.NewFromString(validGasLimit)
	typedGasPrice, _ := amount.NewFromString(validGasPrice)
	typedMaxPriorityFeePerGas, _ := amount.NewFromString(validMaxPriorityFeePerGas)
	typedFeeMultiplier, _ := amount.NewFromString(validFeeMultiplier)
	typedMaxCuPrice, _ := amount.NewFromString("1000000000")

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
		{
			Id:                 validSvmNetworkId,
			Url:                validSvmUrl,
			Dlt:                validSvmDlt,
			GasLimitMultiplier: *amount.Zero(),
			MaxCuPrice:         *typedMaxCuPrice,
		},
	}
}

func newHandler() (
	*CommandHandler,
	*testbus.EventBusMock,
	*npmock.NonceProviderMock,
	*contracttransactionbuilder.Registry,
	*custodymock.CustodyProviderMock,
	*transanctionsender.Registry,
	*servicemock.CustodyKeyExistMultipleMock,
	*mock2.EvmClientRegistryMock,
	*mock2.EvmClientMock,
	*bbqmock.BoundedBlockingQueueMock,
	*mockevmtransactionrepo.Postgres,
	*mocksvmtransactionrepo.Postgres,
	*svmmocks.BlockhashProviderMock,
	*svmmocks.SvmClientRegistryMock,
) {
	eventBus := new(testbus.EventBusMock)
	nonceProvider := new(npmock.NonceProviderMock)
	blockhashProvider := new(svmmocks.BlockhashProviderMock)
	registry := contracttransactionbuilder.NewRegistry()
	custodyProvider := new(custodymock.CustodyProviderMock)
	transactionSenderRegistry := transanctionsender.NewRegistry()
	transactionGasEstimatorRegistry := transactiongasestimator.NewRegistry()
	custodyKeyExistMultipleService := new(servicemock.CustodyKeyExistMultipleMock)
	ethClientRegistry := new(mock2.EvmClientRegistryMock)
	svmClientRegistry := new(svmmocks.SvmClientRegistryMock)
	ethClient := new(mock2.EvmClientMock)
	bbq := new(bbqmock.BoundedBlockingQueueMock)
	evmTransactionRepository := new(mockevmtransactionrepo.Postgres)
	svmTransactionRepository := new(mocksvmtransactionrepo.Postgres)

	txSvc := txservice.NewAppService(nonceProvider, blockhashProvider, *registry, ethClientRegistry, svmClientRegistry, *transactionGasEstimatorRegistry)

	handler := NewCommandHandler(
		eventBus,
		txSvc,
		nonceProvider,
		custodyProvider,
		*transactionSenderRegistry,
		custodyKeyExistMultipleService,
		bbq,
		evmTransactionRepository,
		svmTransactionRepository,
	)

	return handler, eventBus, nonceProvider, registry, custodyProvider, transactionSenderRegistry, custodyKeyExistMultipleService, ethClientRegistry, ethClient, bbq, evmTransactionRepository, svmTransactionRepository, blockhashProvider, svmClientRegistry
}

type expectedEvmTransactionSentEvent struct {
	TxId                 string
	NetworkId            string
	FromAddress          string
	ToAddress            string
	Nonce                amount.Amount
	Value                *amount.Amount
	Data                 string
	TransactionType      uint
	GasLimit             amount.Amount
	GasPrice             amount.Amount
	MaxPriorityFeePerGas amount.Amount
	MaxFeePerGas         amount.Amount
}

type expectedSvmTransactionSentEvent struct {
	TxId                  string
	NetworkId             string
	FeePayer              string
	RecentBlockhash       string
	SerializedTransaction string
}

func matchesEvmTransactionSentEvent(t *testing.T, expected expectedEvmTransactionSentEvent) any {
	return mock.MatchedBy(func(evt any) bool {
		sentEvt, ok := evt.(transaction.TransactionSentEvent)
		if !assert.True(t, ok, "expected transaction.TransactionSentEvent, got %T", evt) {
			return false
		}

		t.Helper()
		assert.NotNil(t, evt)

		assert.Equal(t, expected.TxId, sentEvt.TxId)
		assert.Equal(t, expected.NetworkId, sentEvt.NetworkId)
		assert.Equal(t, validDlt, sentEvt.Dlt)
		assert.Nil(t, sentEvt.SvmTransactionEventModel)

		assert.NotNil(t, sentEvt.EvmTransactionEventModel)
		assert.Equal(t, expected.FromAddress, sentEvt.FromAddress)
		assert.Equal(t, expected.ToAddress, sentEvt.ToAddress)
		assert.True(t, sentEvt.Nonce.Equal(expected.Nonce))
		assert.Equal(t, expected.Value, sentEvt.Value)
		assert.Equal(t, expected.Data, sentEvt.Data)
		assert.Equal(t, expected.TransactionType, sentEvt.TransactionType)
		assert.True(t, sentEvt.GasLimit.Equal(expected.GasLimit))
		assert.True(t, sentEvt.GasPrice.Equal(expected.GasPrice))
		assert.True(t, sentEvt.MaxPriorityFeePerGas.Equal(expected.MaxPriorityFeePerGas))
		assert.True(t, sentEvt.MaxFeePerGas.Equal(expected.MaxFeePerGas))

		return true
	})
}

func matchesSvmTransactionSentEvent(t *testing.T, expected expectedSvmTransactionSentEvent) any {
	return mock.MatchedBy(func(evt any) bool {
		sentEvt, ok := evt.(transaction.TransactionSentEvent)
		if !assert.True(t, ok, "expected transaction.TransactionSentEvent, got %T", evt) {
			return false
		}

		t.Helper()
		assert.NotNil(t, evt)

		assert.Equal(t, expected.TxId, sentEvt.TxId)
		assert.Equal(t, expected.NetworkId, sentEvt.NetworkId)
		assert.Equal(t, validSvmDlt, sentEvt.Dlt)
		assert.Nil(t, sentEvt.EvmTransactionEventModel)

		assert.NotNil(t, sentEvt.SvmTransactionEventModel)
		assert.Equal(t, expected.FeePayer, sentEvt.FeePayer)
		assert.Equal(t, expected.RecentBlockhash, sentEvt.RecentBlockhash)
		assert.Equal(t, expected.SerializedTransaction, sentEvt.SerializedTransaction)
		assert.Nil(t, sentEvt.CuLimit)
		assert.Nil(t, sentEvt.CuPrice)

		return true
	})
}

func TestDltIngressSignAndSendCommandHandler_Handle_BbqPutError(t *testing.T) {
	handler, _, _, _, _, _, _, _, _, bbq, _, _, _, _ := newHandler()

	ctx := context.Background()
	cmd := &Command{
		NetworkId: validNetworkId,
	}

	bbq.On("Put", ctx, validNetworkId, mock.Anything).Return(assert.AnError)

	resp, err := handler.Handle(ctx, cmd)

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, assert.AnError)
	bbq.AssertNotCalled(t, "Take", mock.Anything, mock.Anything)
}

func TestDltIngressSignAndSendCommandHandler_Handle_CustodyKeyNotFound_ReleasesQueue(t *testing.T) {
	handler, _, _, _, _, _, custodyKeyExistsService, _, _, bbq, _, _, _, _ := newHandler()

	ctx := context.Background()
	cmd := &Command{
		SenderDltAccountId:   validSenderDltAccountId,
		SignersDltAccountIds: []string{validSenderDltAccountId},
		NetworkId:            validNetworkId,
	}

	bbq.On("Put", ctx, validNetworkId, mock.Anything).Return(nil)
	bbq.On("Take", ctx, validNetworkId).Return(nil)
	custodyKeyExistsService.On("Execute", ctx, []string{validSenderDltAccountId}).Return(nil, assert.AnError)

	resp, err := handler.Handle(ctx, cmd)

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, assert.AnError)
	bbq.AssertCalled(t, "Take", ctx, validNetworkId)
}

func TestDltIngressSignAndSendCommandHandler_Handle_NetworkNotFound(t *testing.T) {
	handler, _, _, _, _, _, _, _, _, _, _, _, _, _ := newHandler()

	ctx := context.Background()
	cmd := &Command{
		NetworkId: "invalid-network",
	}

	resp, err := handler.Handle(ctx, cmd)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, "Entity Network with invalid-network not found", err.Error())
}

func TestDltIngressSignAndSendCommandHandler_Evm_Handle_Success(t *testing.T) {
	handler, eventBus, nonceProvider, registry, custodyProvider, transactionSenderRegistry, custodyKeyExistsService, ethClientRegistry, _, bbq, transactionRepository, _, _, _ := newHandler()

	ctx := context.Background()
	cmd := &Command{
		SenderDltAccountId:   validSenderDltAccountId,
		SignersDltAccountIds: []string{validSenderDltAccountId},
		SmartContractId:      validSmartContractId,
		SmartContractName:    validSmartContractName,
		MethodName:           validMethodName,
		MethodArgs:           map[string]any{},
		NetworkId:            validNetworkId,
	}

	bbq.On("Put", ctx, validNetworkId, mock.Anything).Return(nil)

	custodyKey := &custodykey.CustodyKey{
		DltAccountId: validSenderDltAccountId,
		ExternalId:   validExternalId,
	}
	custodyKeyExistsService.On("Execute", ctx, []string{validSenderDltAccountId}).Return([]*custodykey.CustodyKey{custodyKey}, nil)

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

	gasLimit, _ := amount.NewFromString(validGasLimit)
	gasPrice, _ := amount.NewFromString(validGasPrice)
	maxPriorityFeePerGas, _ := amount.NewFromString(validMaxPriorityFeePerGas)
	maxFeePerGas, _ := amount.NewFromString(validMaxFeePerGas)
	txResponse := &portcommon.TransactionResponse{
		EVMTransactionResponse: &portcommon.EVMTransactionResponse{
			Data:                 "0x",
			To:                   validSmartContractId,
			Nonce:                nonce,
			TransactionType:      portcommon.TransactionTypeDynamicFee,
			GasLimit:             gasLimit,
			GasPrice:             gasPrice,
			MaxPriorityFeePerGas: maxPriorityFeePerGas,
			MaxFeePerGas:         maxFeePerGas,
			Value:                nil,
		},
	}
	builder.On("BuildTransaction", mock.Anything).Return(txResponse, nil)

	custodyProvider.On("Sign", ctx, &custody.SignRequest{
		Dlt:         validDlt,
		CustodyKeys: []*custodykey.CustodyKey{custodyKey},
		Transaction: txResponse,
	}).Return(&custody.SignResponse{SignedTransaction: "0xabc", TxId: validTxId}, nil)

	transactionSender := new(tsmock.TransactionSenderMock)
	transactionSenderRegistry.Register(common.EVM, transactionSender)

	transactionSender.On("SendTransaction", ctx, transanctionsender.SendTransactionRequest{
		SignedTransaction: "0xabc",
		Dlt:               validDlt,
		NetworkId:         validNetworkId,
	}).Return(&transanctionsender.SendTransactionResponse{TxId: validTxId}, nil)

	transactionRepository.On("Save", mock.Anything, mock.Anything).Return(nil)

	nonceProvider.On("SetNonce", ctx, mock.AnythingOfType("*nonceprovider.SetNonceRequest")).Return(nil)

	eventBus.On("Publish", mock.Anything, matchesEvmTransactionSentEvent(t, expectedEvmTransactionSentEvent{
		TxId:                 validTxId,
		NetworkId:            cmd.NetworkId,
		FromAddress:          cmd.SenderDltAccountId,
		ToAddress:            cmd.SmartContractId,
		Nonce:                *nonce,
		Value:                nil,
		Data:                 "0x",
		TransactionType:      uint(validTransactionType),
		GasLimit:             *gasLimit,
		GasPrice:             *gasPrice,
		MaxPriorityFeePerGas: *maxPriorityFeePerGas,
		MaxFeePerGas:         *maxFeePerGas,
	})).Return(nil)

	resp, err := handler.Handle(ctx, cmd)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, validTxId, resp.TxId)

	bbq.AssertNotCalled(t, "Take", mock.Anything, mock.Anything)
	custodyKeyExistsService.AssertExpectations(t)
	nonceProvider.AssertExpectations(t)
	builder.AssertExpectations(t)
	custodyProvider.AssertExpectations(t)
	transactionSender.AssertExpectations(t)
	ethClientRegistry.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestDltIngressSignAndSendCommandHandler_Evm_Handle_DynamicFee_UsesCalculatedMaxFee(t *testing.T) {
	handler, eventBus, nonceProvider, registry, custodyProvider, transactionSenderRegistry, custodyKeyExistsService, ethClientRegistry, ethClient, bbq, transactionRepository, _, _, _ := newHandler()

	network := config.DltIngressConfig.DltIngress.Networks[0]
	originalType := network.TransactionType
	network.TransactionType = portcommon.TransactionTypeDynamicFee
	t.Cleanup(func() { network.TransactionType = originalType })

	ctx := context.Background()
	cmd := &Command{
		SenderDltAccountId:   validSenderDltAccountId,
		SignersDltAccountIds: []string{validSenderDltAccountId},
		SmartContractId:      validSmartContractId,
		SmartContractName:    validSmartContractName,
		MethodName:           validMethodName,
		MethodArgs:           map[string]any{},
		NetworkId:            validNetworkId,
	}

	bbq.On("Put", ctx, validNetworkId, mock.Anything).Return(nil)

	custodyKey := &custodykey.CustodyKey{
		DltAccountId: validSenderDltAccountId,
		ExternalId:   validExternalId,
	}
	custodyKeyExistsService.On("Execute", ctx, []string{validSenderDltAccountId}).Return([]*custodykey.CustodyKey{custodyKey}, nil)

	nonce, _ := amount.NewFromString("0")
	nonceProvider.On("GetNonce", ctx, &nonceprovider.GetNonceRequest{
		NetworkId:    validNetworkId,
		DltAccountId: validSenderDltAccountId,
	}).Return(nonce, nil)

	baseFee, _ := amount.NewFromString("1000000000")
	feeMultiplier, _ := amount.NewFromString(validFeeMultiplier)
	priorityFee, _ := amount.NewFromString(validMaxPriorityFeePerGas)
	incrementedBaseFee := baseFee.Mul(*feeMultiplier)
	expectedMaxFeePerGas := incrementedBaseFee.Add(*priorityFee)

	ethClientRegistry.On("GetClientForNetworkId", ctx, validNetworkId).Return(ethClient, nil)
	ethClient.On("GetBaseFeePerGas", ctx).Return(baseFee, nil)

	builder := new(ctbmock.ContractTransactionBuilderMock)
	registry.Register(common.EVM, validSmartContractName, builder)

	gasLimit, _ := amount.NewFromString(validGasLimit)
	gasPrice, _ := amount.NewFromString(validGasPrice)
	txResponse := &portcommon.TransactionResponse{
		EVMTransactionResponse: &portcommon.EVMTransactionResponse{
			Data:                 "0x",
			To:                   validSmartContractId,
			Nonce:                nonce,
			TransactionType:      portcommon.TransactionTypeDynamicFee,
			GasLimit:             gasLimit,
			GasPrice:             gasPrice,
			MaxPriorityFeePerGas: priorityFee,
			MaxFeePerGas:         &expectedMaxFeePerGas,
			Value:                nil,
		},
	}
	builder.On("BuildTransaction", mock.MatchedBy(func(req *contracttransactionbuilder.BuildTransactionRequest) bool {
		return req.EVMBuildTransactionRequest != nil &&
			req.EVMBuildTransactionRequest.MaxFeePerGas != nil &&
			req.EVMBuildTransactionRequest.MaxFeePerGas.Equal(expectedMaxFeePerGas)
	})).Return(txResponse, nil)

	custodyProvider.On("Sign", ctx, mock.Anything).
		Return(&custody.SignResponse{SignedTransaction: "0xabc", TxId: validTxId}, nil)

	transactionSender := new(tsmock.TransactionSenderMock)
	transactionSenderRegistry.Register(common.EVM, transactionSender)

	transactionSender.On("SendTransaction", ctx, mock.Anything).
		Return(&transanctionsender.SendTransactionResponse{TxId: validTxId}, nil)

	transactionRepository.On("Save", mock.Anything, mock.Anything).Return(nil)
	nonceProvider.On("SetNonce", ctx, mock.AnythingOfType("*nonceprovider.SetNonceRequest")).Return(nil)

	eventBus.On("Publish", mock.Anything, matchesEvmTransactionSentEvent(t, expectedEvmTransactionSentEvent{
		TxId:                 validTxId,
		NetworkId:            cmd.NetworkId,
		FromAddress:          cmd.SenderDltAccountId,
		ToAddress:            cmd.SmartContractId,
		Nonce:                *nonce,
		Value:                nil,
		Data:                 "0x",
		TransactionType:      portcommon.TransactionTypeDynamicFee,
		GasLimit:             *gasLimit,
		GasPrice:             *gasPrice,
		MaxPriorityFeePerGas: *priorityFee,
		MaxFeePerGas:         expectedMaxFeePerGas,
	})).Return(nil)

	resp, err := handler.Handle(ctx, cmd)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.MaxFeePerGas.Equal(expectedMaxFeePerGas))

	ethClientRegistry.AssertExpectations(t)
	ethClient.AssertExpectations(t)
	builder.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestDltIngressSignAndSendCommandHandler_Evm_Handle_NoTransactionSenderRegistered_ReleasesQueue(t *testing.T) {
	handler, _, _, _, _, _, custodyKeyExistsService, _, _, bbq, _, _, _, _ := newHandler()

	ctx := context.Background()
	cmd := &Command{
		SenderDltAccountId:   validSenderDltAccountId,
		SignersDltAccountIds: []string{validSenderDltAccountId},
		NetworkId:            validNetworkId,
	}

	bbq.On("Put", ctx, validNetworkId, mock.Anything).Return(nil)
	bbq.On("Take", ctx, validNetworkId).Return(nil)
	custodyKeyExistsService.On("Execute", ctx, []string{validSenderDltAccountId}).Return([]*custodykey.CustodyKey{{
		DltAccountId: validSenderDltAccountId,
		ExternalId:   validExternalId,
	}}, nil)

	resp, err := handler.Handle(ctx, cmd)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	bbq.AssertCalled(t, "Take", ctx, validNetworkId)
}

func TestDltIngressSignAndSendCommandHandler_Evm_Handle_PrepareTransactionError_ReleasesQueue(t *testing.T) {
	handler, _, _, _, _, transactionSenderRegistry, custodyKeyExistsService, _, _, bbq, _, _, _, _ := newHandler()
	// No builder registered for EVM, so PrepareTransaction will fail at GetContractTransactionBuilder.

	ctx := context.Background()
	cmd := &Command{
		SenderDltAccountId:   validSenderDltAccountId,
		SignersDltAccountIds: []string{validSenderDltAccountId},
		SmartContractId:      validSmartContractId,
		SmartContractName:    validSmartContractName,
		MethodName:           validMethodName,
		MethodArgs:           map[string]any{},
		NetworkId:            validNetworkId,
	}

	bbq.On("Put", ctx, validNetworkId, mock.Anything).Return(nil)
	bbq.On("Take", ctx, validNetworkId).Return(nil)
	custodyKeyExistsService.On("Execute", ctx, []string{validSenderDltAccountId}).Return([]*custodykey.CustodyKey{{
		DltAccountId: validSenderDltAccountId,
		ExternalId:   validExternalId,
	}}, nil)
	transactionSenderRegistry.Register(common.EVM, new(tsmock.TransactionSenderMock))

	resp, err := handler.Handle(ctx, cmd)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	bbq.AssertCalled(t, "Take", ctx, validNetworkId)
}

func TestDltIngressSignAndSendCommandHandler_Svm_Handle_NoTransactionSenderRegistered_ReleasesQueue(t *testing.T) {
	handler, _, _, _, _, _, custodyKeyExistsService, _, _, bbq, _, _, _, _ := newHandler()

	ctx := context.Background()
	cmd := &Command{
		SenderDltAccountId:   validSvmSender,
		SignersDltAccountIds: []string{validSvmSender},
		NetworkId:            validSvmNetworkId,
	}

	bbq.On("Put", ctx, validSvmNetworkId, mock.Anything).Return(nil)
	bbq.On("Take", ctx, validSvmNetworkId).Return(nil)
	custodyKeyExistsService.On("Execute", ctx, []string{validSvmSender}).Return([]*custodykey.CustodyKey{{
		DltAccountId: validSvmSender,
		ExternalId:   validExternalId,
	}}, nil)

	resp, err := handler.Handle(ctx, cmd)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	bbq.AssertCalled(t, "Take", ctx, validSvmNetworkId)
}

func TestDltIngressSignAndSendCommandHandler_Svm_Handle_PrepareTransactionError_ReleasesQueue(t *testing.T) {
	handler, _, _, _, _, transactionSenderRegistry, custodyKeyExistsService, _, _, bbq, _, _, _, _ := newHandler()
	// No builder registered for SVM, so PrepareTransaction will fail at GetContractTransactionBuilder.

	ctx := context.Background()
	cmd := &Command{
		SenderDltAccountId:   validSvmSender,
		SignersDltAccountIds: []string{validSvmSender},
		SmartContractId:      validSvmContractId,
		SmartContractName:    validSmartContractName,
		MethodName:           validMethodName,
		MethodArgs:           map[string]any{},
		NetworkId:            validSvmNetworkId,
	}

	bbq.On("Put", ctx, validSvmNetworkId, mock.Anything).Return(nil)
	bbq.On("Take", ctx, validSvmNetworkId).Return(nil)
	custodyKeyExistsService.On("Execute", ctx, []string{validSvmSender}).Return([]*custodykey.CustodyKey{{
		DltAccountId: validSvmSender,
		ExternalId:   validExternalId,
	}}, nil)
	transactionSenderRegistry.Register(common.SVM, new(tsmock.TransactionSenderMock))

	resp, err := handler.Handle(ctx, cmd)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	bbq.AssertCalled(t, "Take", ctx, validSvmNetworkId)
}

func TestDltIngressSignAndSendCommandHandler_Evm_Handle_SignError_ReleasesQueue(t *testing.T) {
	handler, _, nonceProvider, registry, custodyProvider, transactionSenderRegistry, custodyKeyExistsService, ethClientRegistry, _, bbq, _, _, _, _ := newHandler()

	ctx := context.Background()
	cmd := &Command{
		SenderDltAccountId:   validSenderDltAccountId,
		SignersDltAccountIds: []string{validSenderDltAccountId},
		SmartContractId:      validSmartContractId,
		SmartContractName:    validSmartContractName,
		MethodName:           validMethodName,
		MethodArgs:           map[string]any{},
		NetworkId:            validNetworkId,
	}

	bbq.On("Put", ctx, validNetworkId, mock.Anything).Return(nil)
	bbq.On("Take", ctx, validNetworkId).Return(nil)

	custodyKeyExistsService.On("Execute", ctx, []string{validSenderDltAccountId}).Return([]*custodykey.CustodyKey{{
		DltAccountId: validSenderDltAccountId,
		ExternalId:   validExternalId,
	}}, nil)

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
	builder.On("BuildTransaction", mock.Anything).Return(&portcommon.TransactionResponse{
		EVMTransactionResponse: &portcommon.EVMTransactionResponse{
			Data:  "0x",
			To:    validSmartContractId,
			Nonce: nonce,
		},
	}, nil)

	transactionSenderRegistry.Register(common.EVM, new(tsmock.TransactionSenderMock))
	custodyProvider.On("Sign", ctx, mock.Anything).Return(nil, assert.AnError)

	resp, err := handler.Handle(ctx, cmd)

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, assert.AnError)
	bbq.AssertCalled(t, "Take", ctx, validNetworkId)
}

func TestDltIngressSignAndSendCommandHandler_Evm_Handle_SendError_ReleasesQueue(t *testing.T) {
	handler, _, nonceProvider, registry, custodyProvider, transactionSenderRegistry, custodyKeyExistsService, ethClientRegistry, _, bbq, transactionRepository, _, _, _ := newHandler()

	ctx := context.Background()
	cmd := &Command{
		SenderDltAccountId:   validSenderDltAccountId,
		SignersDltAccountIds: []string{validSenderDltAccountId},
		SmartContractId:      validSmartContractId,
		SmartContractName:    validSmartContractName,
		MethodName:           validMethodName,
		MethodArgs:           map[string]any{},
		NetworkId:            validNetworkId,
	}

	bbq.On("Put", ctx, validNetworkId, mock.Anything).Return(nil)
	bbq.On("Take", ctx, validNetworkId).Return(nil)

	custodyKeyExistsService.On("Execute", ctx, []string{validSenderDltAccountId}).Return([]*custodykey.CustodyKey{{
		DltAccountId: validSenderDltAccountId,
		ExternalId:   validExternalId,
	}}, nil)

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
	builder.On("BuildTransaction", mock.Anything).Return(&portcommon.TransactionResponse{
		EVMTransactionResponse: &portcommon.EVMTransactionResponse{
			Data:  "0x",
			To:    validSmartContractId,
			Nonce: nonce,
		},
	}, nil)

	custodyProvider.On("Sign", ctx, mock.Anything).Return(&custody.SignResponse{SignedTransaction: "0xabc", TxId: validTxId}, nil)

	transactionSender := new(tsmock.TransactionSenderMock)
	transactionSenderRegistry.Register(common.EVM, transactionSender)
	transactionSender.On("SendTransaction", ctx, mock.Anything).Return(nil, assert.AnError)

	transactionRepository.On("Save", mock.Anything, mock.Anything).Return(nil)
	transactionRepository.On("HardDelete", mock.Anything, mock.Anything).Return(nil)

	resp, err := handler.Handle(ctx, cmd)

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, assert.AnError)
	bbq.AssertCalled(t, "Take", ctx, validNetworkId)
	transactionRepository.AssertExpectations(t)
}

func TestDltIngressSignAndSendCommandHandler_Evm_Handle_EventPublishError_KeepsQueueSlot(t *testing.T) {
	handler, eventBus, nonceProvider, registry, custodyProvider, transactionSenderRegistry, custodyKeyExistsService, ethClientRegistry, _, bbq, transactionRepository, _, _, _ := newHandler()

	ctx := context.Background()
	cmd := &Command{
		SenderDltAccountId:   validSenderDltAccountId,
		SignersDltAccountIds: []string{validSenderDltAccountId},
		SmartContractId:      validSmartContractId,
		SmartContractName:    validSmartContractName,
		MethodName:           validMethodName,
		MethodArgs:           map[string]any{},
		NetworkId:            validNetworkId,
	}

	bbq.On("Put", ctx, validNetworkId, mock.Anything).Return(nil)

	custodyKeyExistsService.On("Execute", ctx, []string{validSenderDltAccountId}).Return([]*custodykey.CustodyKey{{
		DltAccountId: validSenderDltAccountId,
		ExternalId:   validExternalId,
	}}, nil)

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

	gasLimit, _ := amount.NewFromString(validGasLimit)
	gasPrice, _ := amount.NewFromString(validGasPrice)
	maxPriorityFeePerGas, _ := amount.NewFromString(validMaxPriorityFeePerGas)
	maxFeePerGas, _ := amount.NewFromString(validMaxFeePerGas)
	txResponse := &portcommon.TransactionResponse{
		EVMTransactionResponse: &portcommon.EVMTransactionResponse{
			Data:                 "0x",
			To:                   validSmartContractId,
			Nonce:                nonce,
			TransactionType:      portcommon.TransactionTypeDynamicFee,
			GasLimit:             gasLimit,
			GasPrice:             gasPrice,
			MaxPriorityFeePerGas: maxPriorityFeePerGas,
			MaxFeePerGas:         maxFeePerGas,
		},
	}
	builder.On("BuildTransaction", mock.Anything).Return(txResponse, nil)

	custodyProvider.On("Sign", ctx, mock.Anything).Return(&custody.SignResponse{SignedTransaction: "0xabc", TxId: validTxId}, nil)

	transactionSender := new(tsmock.TransactionSenderMock)
	transactionSenderRegistry.Register(common.EVM, transactionSender)
	transactionSender.On("SendTransaction", ctx, mock.Anything).Return(&transanctionsender.SendTransactionResponse{TxId: validTxId}, nil)

	transactionRepository.On("Save", mock.Anything, mock.Anything).Return(nil)
	nonceProvider.On("SetNonce", ctx, mock.AnythingOfType("*nonceprovider.SetNonceRequest")).Return(nil)
	eventBus.On("Publish", mock.Anything, mock.Anything).Return(assert.AnError)

	resp, err := handler.Handle(ctx, cmd)

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, assert.AnError)
	bbq.AssertNotCalled(t, "Take", mock.Anything, mock.Anything)
}

func TestDltIngressSignAndSendCommandHandler_Evm_Handle_SetNonceError_KeepsQueueSlot(t *testing.T) {
	handler, _, nonceProvider, registry, custodyProvider, transactionSenderRegistry, custodyKeyExistsService, ethClientRegistry, _, bbq, transactionRepository, _, _, _ := newHandler()

	ctx := context.Background()
	cmd := &Command{
		SenderDltAccountId:   validSenderDltAccountId,
		SignersDltAccountIds: []string{validSenderDltAccountId},
		SmartContractId:      validSmartContractId,
		SmartContractName:    validSmartContractName,
		MethodName:           validMethodName,
		MethodArgs:           map[string]any{},
		NetworkId:            validNetworkId,
	}

	bbq.On("Put", ctx, validNetworkId, mock.Anything).Return(nil)

	custodyKeyExistsService.On("Execute", ctx, []string{validSenderDltAccountId}).Return([]*custodykey.CustodyKey{{
		DltAccountId: validSenderDltAccountId,
		ExternalId:   validExternalId,
	}}, nil)

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

	gasLimit, _ := amount.NewFromString(validGasLimit)
	gasPrice, _ := amount.NewFromString(validGasPrice)
	maxPriorityFeePerGas, _ := amount.NewFromString(validMaxPriorityFeePerGas)
	maxFeePerGas, _ := amount.NewFromString(validMaxFeePerGas)
	txResponse := &portcommon.TransactionResponse{
		EVMTransactionResponse: &portcommon.EVMTransactionResponse{
			Data:                 "0x",
			To:                   validSmartContractId,
			Nonce:                nonce,
			TransactionType:      portcommon.TransactionTypeDynamicFee,
			GasLimit:             gasLimit,
			GasPrice:             gasPrice,
			MaxPriorityFeePerGas: maxPriorityFeePerGas,
			MaxFeePerGas:         maxFeePerGas,
		},
	}
	builder.On("BuildTransaction", mock.Anything).Return(txResponse, nil)

	custodyProvider.On("Sign", ctx, mock.Anything).Return(&custody.SignResponse{SignedTransaction: "0xabc", TxId: validTxId}, nil)

	transactionSender := new(tsmock.TransactionSenderMock)
	transactionSenderRegistry.Register(common.EVM, transactionSender)
	transactionSender.On("SendTransaction", ctx, mock.Anything).Return(&transanctionsender.SendTransactionResponse{TxId: validTxId}, nil)

	transactionRepository.On("Save", mock.Anything, mock.Anything).Return(nil)
	nonceProvider.On("SetNonce", ctx, mock.AnythingOfType("*nonceprovider.SetNonceRequest")).Return(assert.AnError)

	resp, err := handler.Handle(ctx, cmd)

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, assert.AnError)
	bbq.AssertNotCalled(t, "Take", mock.Anything, mock.Anything)
}

func TestDltIngressSignAndSendCommandHandler_Svm_Handle_Success(t *testing.T) {
	handler, eventBus, _, registry, custodyProvider, transactionSenderRegistry, custodyKeyExistsService, _, _, bbq, _, svmTransactionRepository, blockhashProvider, svmClientRegistry := newHandler()

	ctx := context.Background()
	cmd := &Command{
		SenderDltAccountId:   validSvmSender,
		SignersDltAccountIds: []string{validSvmSender},
		SmartContractId:      validSvmContractId,
		SmartContractName:    validSmartContractName,
		MethodName:           validMethodName,
		MethodArgs:           map[string]any{},
		NetworkId:            validSvmNetworkId,
	}

	bbq.On("Put", ctx, validSvmNetworkId, mock.Anything).Return(nil)

	custodyKey := &custodykey.CustodyKey{
		DltAccountId: validSvmSender,
		ExternalId:   validExternalId,
	}
	custodyKeyExistsService.On("Execute", ctx, []string{validSvmSender}).Return([]*custodykey.CustodyKey{custodyKey}, nil)

	blockhashProvider.On("GetRecentBlockhash", ctx, validSvmNetworkId).Return(validSvmBlockhash, nil)
	svmClient := new(svmmocks.SvmClientMock)
	svmClientRegistry.On("GetClientForNetworkId", validSvmNetworkId).Return(svmClient, nil)
	svmClient.On("GetRecentPrioritizationFees", ctx, []string{}).Return([]uint64{1000, 2000, 3000}, nil)

	svmTxResponse := &portcommon.TransactionResponse{
		SVMTransactionResponse: &portcommon.SVMTransactionResponse{
			SerializedTransaction: validSvmSerializedTx,
			FeePayer:              validSvmSender,
			RecentBlockhash:       validSvmBlockhash,
		},
	}
	builder := new(ctbmock.ContractTransactionBuilderMock)
	registry.Register(common.SVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(svmTxResponse, nil)
	builder.On("OverrideCuPrice", mock.Anything, mock.Anything, mock.Anything).Return(svmTxResponse, nil)

	custodyProvider.On("Sign", ctx, &custody.SignRequest{
		Dlt:         validSvmDlt,
		CustodyKeys: []*custodykey.CustodyKey{custodyKey},
		Transaction: svmTxResponse,
	}).Return(&custody.SignResponse{SignedTransaction: "signedSvmTx", TxId: validSvmTxId}, nil)

	transactionSender := new(tsmock.TransactionSenderMock)
	transactionSenderRegistry.Register(common.SVM, transactionSender)
	transactionSender.On("SendTransaction", ctx, transanctionsender.SendTransactionRequest{
		SignedTransaction: "signedSvmTx",
		Dlt:               validSvmDlt,
		NetworkId:         validSvmNetworkId,
	}).Return(&transanctionsender.SendTransactionResponse{TxId: validSvmTxId}, nil)

	svmTransactionRepository.On("Save", mock.Anything, mock.Anything).Return(nil)

	eventBus.On("Publish", mock.Anything, matchesSvmTransactionSentEvent(t, expectedSvmTransactionSentEvent{
		TxId:                  validSvmTxId,
		NetworkId:             cmd.NetworkId,
		FeePayer:              cmd.SenderDltAccountId,
		RecentBlockhash:       validSvmBlockhash,
		SerializedTransaction: validSvmSerializedTx,
	})).Return(nil)

	resp, err := handler.Handle(ctx, cmd)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, validSvmTxId, resp.TxId)
	assert.Equal(t, validSvmDlt, resp.Dlt)
	assert.Equal(t, validSvmNetworkId, resp.NetworkId)

	bbq.AssertNotCalled(t, "Take", mock.Anything, mock.Anything)
	custodyKeyExistsService.AssertExpectations(t)
	builder.AssertExpectations(t)
	custodyProvider.AssertExpectations(t)
	transactionSender.AssertExpectations(t)
	svmTransactionRepository.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestDltIngressSignAndSendCommandHandler_Svm_Handle_SignError_ReleasesQueue(t *testing.T) {
	handler, _, _, registry, custodyProvider, transactionSenderRegistry, custodyKeyExistsService, _, _, bbq, _, _, blockhashProvider, svmClientRegistry := newHandler()

	ctx := context.Background()
	cmd := &Command{
		SenderDltAccountId:   validSvmSender,
		SignersDltAccountIds: []string{validSvmSender},
		SmartContractId:      validSvmContractId,
		SmartContractName:    validSmartContractName,
		MethodName:           validMethodName,
		MethodArgs:           map[string]any{},
		NetworkId:            validSvmNetworkId,
	}

	bbq.On("Put", ctx, validSvmNetworkId, mock.Anything).Return(nil)
	bbq.On("Take", ctx, validSvmNetworkId).Return(nil)

	custodyKeyExistsService.On("Execute", ctx, []string{validSvmSender}).Return([]*custodykey.CustodyKey{{
		DltAccountId: validSvmSender,
		ExternalId:   validExternalId,
	}}, nil)

	blockhashProvider.On("GetRecentBlockhash", ctx, validSvmNetworkId).Return(validSvmBlockhash, nil)
	svmClient := new(svmmocks.SvmClientMock)
	svmClientRegistry.On("GetClientForNetworkId", validSvmNetworkId).Return(svmClient, nil)
	svmClient.On("GetRecentPrioritizationFees", ctx, []string{}).Return([]uint64{1000}, nil)

	svmTxResponse := &portcommon.TransactionResponse{
		SVMTransactionResponse: &portcommon.SVMTransactionResponse{
			SerializedTransaction: validSvmSerializedTx,
			FeePayer:              validSvmSender,
			RecentBlockhash:       validSvmBlockhash,
		},
	}
	builder := new(ctbmock.ContractTransactionBuilderMock)
	registry.Register(common.SVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(svmTxResponse, nil)
	builder.On("OverrideCuPrice", mock.Anything, mock.Anything, mock.Anything).Return(svmTxResponse, nil)

	transactionSenderRegistry.Register(common.SVM, new(tsmock.TransactionSenderMock))
	custodyProvider.On("Sign", ctx, mock.Anything).Return(nil, assert.AnError)

	resp, err := handler.Handle(ctx, cmd)

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, assert.AnError)
	bbq.AssertCalled(t, "Take", ctx, validSvmNetworkId)
}

func TestDltIngressSignAndSendCommandHandler_Svm_Handle_SendError_ReleasesQueue(t *testing.T) {
	handler, _, _, registry, custodyProvider, transactionSenderRegistry, custodyKeyExistMultipleService, _, _, bbq, _, svmTransactionRepository, blockhashProvider, svmClientRegistry := newHandler()

	ctx := context.Background()
	cmd := &Command{
		SenderDltAccountId:   validSvmSender,
		SignersDltAccountIds: []string{validSvmSender},
		SmartContractId:      validSvmContractId,
		SmartContractName:    validSmartContractName,
		MethodName:           validMethodName,
		MethodArgs:           map[string]any{},
		NetworkId:            validSvmNetworkId,
	}

	bbq.On("Put", ctx, validSvmNetworkId, mock.Anything).Return(nil)
	bbq.On("Take", ctx, validSvmNetworkId).Return(nil)

	custodyKeyExistMultipleService.On("Execute", ctx, []string{validSvmSender}).Return([]*custodykey.CustodyKey{{
		DltAccountId: validSvmSender,
		ExternalId:   validExternalId,
	}}, nil)

	blockhashProvider.On("GetRecentBlockhash", ctx, validSvmNetworkId).Return(validSvmBlockhash, nil)
	svmClient := new(svmmocks.SvmClientMock)
	svmClientRegistry.On("GetClientForNetworkId", validSvmNetworkId).Return(svmClient, nil)
	svmClient.On("GetRecentPrioritizationFees", ctx, []string{}).Return([]uint64{1000}, nil)

	svmTxResponse := &portcommon.TransactionResponse{
		SVMTransactionResponse: &portcommon.SVMTransactionResponse{
			SerializedTransaction: validSvmSerializedTx,
			FeePayer:              validSvmSender,
			RecentBlockhash:       validSvmBlockhash,
		},
	}
	builder := new(ctbmock.ContractTransactionBuilderMock)
	registry.Register(common.SVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(svmTxResponse, nil)
	builder.On("OverrideCuPrice", mock.Anything, mock.Anything, mock.Anything).Return(svmTxResponse, nil)

	custodyProvider.On("Sign", ctx, mock.Anything).Return(&custody.SignResponse{SignedTransaction: "signedSvmTx", TxId: validSvmTxId}, nil)

	transactionSender := new(tsmock.TransactionSenderMock)
	transactionSenderRegistry.Register(common.SVM, transactionSender)
	transactionSender.On("SendTransaction", ctx, mock.Anything).Return(nil, assert.AnError)

	svmTransactionRepository.On("Save", mock.Anything, mock.Anything).Return(nil)
	svmTransactionRepository.On("HardDelete", mock.Anything, mock.Anything).Return(nil)

	resp, err := handler.Handle(ctx, cmd)

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, assert.AnError)
	bbq.AssertCalled(t, "Take", ctx, validSvmNetworkId)
	svmTransactionRepository.AssertExpectations(t)
}

func TestDltIngressSignAndSendCommandHandler_Svm_Handle_EventPublishError_KeepsQueueSlot(t *testing.T) {
	handler, eventBus, _, registry, custodyProvider, transactionSenderRegistry, custodyKeyExistsService, _, _, bbq, _, svmTransactionRepository, blockhashProvider, svmClientRegistry := newHandler()

	ctx := context.Background()
	cmd := &Command{
		SenderDltAccountId:   validSvmSender,
		SignersDltAccountIds: []string{validSvmSender},
		SmartContractId:      validSvmContractId,
		SmartContractName:    validSmartContractName,
		MethodName:           validMethodName,
		MethodArgs:           map[string]any{},
		NetworkId:            validSvmNetworkId,
	}

	bbq.On("Put", ctx, validSvmNetworkId, mock.Anything).Return(nil)

	custodyKeyExistsService.On("Execute", ctx, []string{validSvmSender}).Return([]*custodykey.CustodyKey{{
		DltAccountId: validSvmSender,
		ExternalId:   validExternalId,
	}}, nil)

	blockhashProvider.On("GetRecentBlockhash", ctx, validSvmNetworkId).Return(validSvmBlockhash, nil)
	svmClient := new(svmmocks.SvmClientMock)
	svmClientRegistry.On("GetClientForNetworkId", validSvmNetworkId).Return(svmClient, nil)
	svmClient.On("GetRecentPrioritizationFees", ctx, []string{}).Return([]uint64{1000, 2000, 3000}, nil)

	svmTxResponse := &portcommon.TransactionResponse{
		SVMTransactionResponse: &portcommon.SVMTransactionResponse{
			SerializedTransaction: validSvmSerializedTx,
			FeePayer:              validSvmSender,
			RecentBlockhash:       validSvmBlockhash,
		},
	}
	builder := new(ctbmock.ContractTransactionBuilderMock)
	registry.Register(common.SVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).Return(svmTxResponse, nil)
	builder.On("OverrideCuPrice", mock.Anything, mock.Anything, mock.Anything).Return(svmTxResponse, nil)

	custodyProvider.On("Sign", ctx, mock.Anything).Return(&custody.SignResponse{SignedTransaction: "signedSvmTx", TxId: validSvmTxId}, nil)

	transactionSender := new(tsmock.TransactionSenderMock)
	transactionSenderRegistry.Register(common.SVM, transactionSender)
	transactionSender.On("SendTransaction", ctx, mock.Anything).Return(&transanctionsender.SendTransactionResponse{TxId: validSvmTxId}, nil)

	svmTransactionRepository.On("Save", mock.Anything, mock.Anything).Return(nil)
	eventBus.On("Publish", mock.Anything, mock.Anything).Return(assert.AnError)

	resp, err := handler.Handle(ctx, cmd)

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, assert.AnError)
	bbq.AssertNotCalled(t, "Take", mock.Anything, mock.Anything)
}
