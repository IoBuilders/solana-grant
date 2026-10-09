package retryevmtransaction

import (
	"context"
	"testing"

	"dlt-ingress/src/main/dltingress/internal/domain/custodykey"
	custodykeyfactory "dlt-ingress/src/main/dltingress/internal/domain/custodykey"
	servicemock "dlt-ingress/src/main/dltingress/internal/domain/custodykey/service/mock"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction/evmtransaction"
	bbqmock "dlt-ingress/src/main/dltingress/internal/infra/boundedblockingqueue/mock"
	mock2 "dlt-ingress/src/main/dltingress/internal/infra/evm/mock"
	"dlt-ingress/src/main/dltingress/internal/infra/nonceprovider"
	npmock "dlt-ingress/src/main/dltingress/internal/infra/nonceprovider/mock"
	"dlt-ingress/src/main/dltingress/internal/infra/repository/transaction/evmtransaction/mock"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction"
	txfactory "dlt-ingress/src/main/dltingress/internal/domain/transaction/evmtransaction"
	"dlt-ingress/src/main/dltingress/internal/infra/custody"
	custodymock "dlt-ingress/src/main/dltingress/internal/infra/custody/mock"
	"dlt-ingress/src/main/dltingress/internal/infra/portcommon"
	"dlt-ingress/src/main/dltingress/internal/infra/transanctionsender"
	tsmock "dlt-ingress/src/main/dltingress/internal/infra/transanctionsender/mock"
	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	testbus "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/test/core/event"
)

const (
	validNetworkId = "1"
	validDlt       = "EVM"
	validUrl       = "http://localhost:8545"
	validChainId   = "1"
	// gasLimit is both the transaction seed and the ceiling a client-supplied override may
	// not exceed, so the fixture carries the real Hedera ceiling.
	validGasLimit             = "15000000"
	validMaxPriorityFeePerGas = "1000000000"
	validFeeMultiplier        = "1.2"
	validTxId                 = "0x789"
	validTransactionType      = 0
	validGasPrice             = "1000000000"
	validMaxTxPoolSize        = 100
)

type testContext struct {
	handler                   *CommandHandler
	eventBus                  *testbus.EventBusMock
	custodyProvider           *custodymock.CustodyProviderMock
	nonceProvider             *npmock.NonceProviderMock
	transactionSenderRegistry *transanctionsender.Registry
	ethClientRegistry         *mock2.EvmClientRegistryMock
	ethClient                 *mock2.EvmClientMock
	transactionRepository     *mockevmtransactionrepo.Postgres
	boundedBlockingQueue      *bbqmock.BoundedBlockingQueueMock
	custodyKeyExistsService   *servicemock.CustodyKeyExistsMock
}

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
		},
	}
}

func newTestContext() *testContext {
	eventBus := new(testbus.EventBusMock)
	ethClientRegistry := new(mock2.EvmClientRegistryMock)
	ethClient := new(mock2.EvmClientMock)
	custodyProvider := new(custodymock.CustodyProviderMock)
	nonceProvider := new(npmock.NonceProviderMock)
	transactionSenderRegistry := transanctionsender.NewRegistry()
	transactionRepository := new(mockevmtransactionrepo.Postgres)
	bbq := new(bbqmock.BoundedBlockingQueueMock)
	custodyKeyExistsService := new(servicemock.CustodyKeyExistsMock)

	handler := NewCommandHandler(
		eventBus,
		custodyProvider,
		nonceProvider,
		*transactionSenderRegistry,
		transactionRepository,
		ethClientRegistry,
		bbq,
		custodyKeyExistsService,
	)

	return &testContext{
		handler:                   handler,
		eventBus:                  eventBus,
		custodyProvider:           custodyProvider,
		nonceProvider:             nonceProvider,
		transactionSenderRegistry: transactionSenderRegistry,
		ethClientRegistry:         ethClientRegistry,
		ethClient:                 ethClient,
		transactionRepository:     transactionRepository,
		boundedBlockingQueue:      bbq,
		custodyKeyExistsService:   custodyKeyExistsService,
	}
}

func registerSender(tc *testContext, ctx context.Context, tx *evmtransaction.EvmTransaction) *tsmock.TransactionSenderMock {
	transactionSender := new(tsmock.TransactionSenderMock)
	tc.transactionSenderRegistry.Register(common.EVM, transactionSender)

	transactionSender.
		On("SendTransaction", ctx, transanctionsender.SendTransactionRequest{
			SignedTransaction: "0xabc",
			Dlt:               string(tx.Dlt),
			NetworkId:         tx.NetworkId,
		}).
		Return(&transanctionsender.SendTransactionResponse{TxId: validTxId}, nil)

	return transactionSender
}

func expectFindByTxId(tc *testContext, ctx context.Context, tx *evmtransaction.EvmTransaction) {
	tc.transactionRepository.On("FindByTxId", ctx, tx.TxId).Return(tx, nil)
}

func expectCustodyKeyExists(tc *testContext, ctx context.Context, sender string, custodyKey *custodykey.CustodyKey) {
	tc.custodyKeyExistsService.On("Execute", ctx, sender).Return(custodyKey, nil)
}

func expectSave(tc *testContext, ctx context.Context, matcher func(saved *evmtransaction.EvmTransaction) bool) {
	tc.transactionRepository.
		On("Save", ctx, mock.MatchedBy(matcher)).
		Return(nil)
}

func expectPublish(tc *testContext, ctx context.Context, oldTxId string) {
	tc.eventBus.
		On("Publish", ctx, mock.MatchedBy(func(evt transaction.RetriedEvent) bool {
			return evt.NewTxId == validTxId && evt.OldTxId == oldTxId
		})).
		Return(nil)
}

func expectCalculateMaxFee(
	tc *testContext,
	ctx context.Context,
	tx *evmtransaction.EvmTransaction,
	priorityFee *amount.Amount,
) amount.Amount {
	network, _ := config.DltIngressConfig.DltIngress.GetDltIngressNetwork(tx.NetworkId)
	baseFee, _ := amount.NewFromString("1000000000")
	tc.ethClientRegistry.On("GetClientForNetworkId", ctx, tx.NetworkId).Return(tc.ethClient, nil)
	tc.ethClient.On("GetBaseFeePerGas", ctx).Return(baseFee, nil)
	incrementedBaseFee := baseFee.MulWithDecimals(network.FeeMultiplier, 0)
	return incrementedBaseFee.Add(*priorityFee)
}

func buildExpectedSignRequest(
	tx *evmtransaction.EvmTransaction,
	key *custodykey.CustodyKey,
	chainID *amount.Amount,
	nonce *amount.Amount,
	gasLimit *amount.Amount,
	gasPrice *amount.Amount,
	maxPriorityFeePerGas *amount.Amount,
	maxFeePerGas *amount.Amount,
) *custody.SignRequest {
	return &custody.SignRequest{
		Dlt:         string(tx.Dlt),
		CustodyKeys: []*custodykey.CustodyKey{key},
		Transaction: &portcommon.TransactionResponse{
			EVMTransactionResponse: &portcommon.EVMTransactionResponse{
				TransactionType:      tx.TransactionType,
				ChainId:              chainID,
				Nonce:                nonce,
				GasLimit:             gasLimit,
				GasPrice:             gasPrice,
				MaxPriorityFeePerGas: maxPriorityFeePerGas,
				MaxFeePerGas:         maxFeePerGas,
				Data:                 tx.Data,
				Value:                tx.Value,
				To:                   tx.ToAddress,
			},
		},
	}
}

func expectSign(
	tc *testContext,
	ctx context.Context,
	req *custody.SignRequest,
) {
	tc.custodyProvider.
		On("Sign", ctx, req).
		Return(&custody.SignResponse{SignedTransaction: "0xabc", TxId: validTxId}, nil)
}

func TestDltIngressRetryTransactionCommandHandler_TxTypeLegacy_Handle_Success(t *testing.T) {
	factory := txfactory.NewEvmTransactionTestFactory()
	custodyKeyFactory := custodykeyfactory.NewCustodyKeyTestFactory()
	typedChainId, _ := amount.NewFromString(validChainId)

	t.Run("same GasPrice and GasLimit", func(t *testing.T) {
		tx := factory.CreateEntity(func(t *evmtransaction.EvmTransaction) {
			t.TransactionType = portcommon.TransactionTypeLegacy
		})
		custodyKey := custodyKeyFactory.CreateEntity(func(key *custodykey.CustodyKey) {
			key.DltAccountId = tx.FromAddress
		})
		tc := newTestContext()
		ctx := context.Background()
		cmd := &Command{TransactionHash: tx.TxId}

		tc.boundedBlockingQueue.On("Put", ctx, validNetworkId, mock.Anything).Return(nil)

		newNonce, _ := amount.NewFromString("10")
		tc.nonceProvider.On("GetNonce", ctx, &nonceprovider.GetNonceRequest{
			NetworkId:    tx.NetworkId,
			DltAccountId: tx.FromAddress,
		}).Return(newNonce, nil)

		expectFindByTxId(tc, ctx, tx)
		expectCustodyKeyExists(tc, ctx, tx.FromAddress, custodyKey)
		expectSign(tc, ctx, buildExpectedSignRequest(
			tx,
			custodyKey,
			typedChainId,
			newNonce,
			tx.GasLimit,
			tx.GasPrice,
			tx.MaxPriorityFeePerGas,
			tx.MaxFeePerGas,
		))
		sender := registerSender(tc, ctx, tx)
		expectSave(tc, ctx, func(saved *evmtransaction.EvmTransaction) bool {
			return saved != nil &&
				saved.TxId == validTxId &&
				saved.GasPrice != nil &&
				saved.GasPrice.Equal(*tx.GasPrice) &&
				saved.OriginalTxId == &tx.TxId
		})
		expectPublish(tc, ctx, cmd.TransactionHash)

		resp, err := tc.handler.Handle(ctx, cmd)

		assert.Nil(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, validTxId, resp.NewTxId)
		sender.AssertExpectations(t)
		tc.transactionRepository.AssertExpectations(t)
		tc.boundedBlockingQueue.AssertExpectations(t)
		tc.nonceProvider.AssertExpectations(t)
		tc.custodyProvider.AssertExpectations(t)
		tc.ethClientRegistry.AssertNotCalled(t, "GetClientForNetworkId", mock.Anything, mock.Anything)
	})

	t.Run("GasPrice override", func(t *testing.T) {
		tx := factory.CreateEntity(func(t *evmtransaction.EvmTransaction) {
			t.TransactionType = portcommon.TransactionTypeLegacy
		})
		custodyKey := custodyKeyFactory.CreateEntity(func(key *custodykey.CustodyKey) {
			key.DltAccountId = tx.FromAddress
		})
		tc := newTestContext()
		ctx := context.Background()
		overrideGasPrice, _ := amount.NewFromString("2000000000")
		cmd := &Command{
			TransactionHash: tx.TxId,
			GasPrice:        overrideGasPrice,
		}

		tc.boundedBlockingQueue.On("Put", ctx, validNetworkId, mock.Anything).Return(nil)

		newNonce, _ := amount.NewFromString("10")
		tc.nonceProvider.On("GetNonce", ctx, &nonceprovider.GetNonceRequest{
			NetworkId:    tx.NetworkId,
			DltAccountId: tx.FromAddress,
		}).Return(newNonce, nil)

		expectFindByTxId(tc, ctx, tx)
		expectCustodyKeyExists(tc, ctx, tx.FromAddress, custodyKey)
		expectSign(tc, ctx, buildExpectedSignRequest(
			tx,
			custodyKey,
			typedChainId,
			newNonce,
			tx.GasLimit,
			overrideGasPrice,
			tx.MaxPriorityFeePerGas,
			tx.MaxFeePerGas,
		))
		sender := registerSender(tc, ctx, tx)
		expectSave(tc, ctx, func(saved *evmtransaction.EvmTransaction) bool {
			return saved != nil &&
				saved.TxId == validTxId &&
				saved.GasPrice != nil &&
				saved.GasPrice.Equal(*overrideGasPrice)
		})
		expectPublish(tc, ctx, cmd.TransactionHash)

		resp, err := tc.handler.Handle(ctx, cmd)

		assert.Nil(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, validTxId, resp.NewTxId)
		sender.AssertExpectations(t)
		tc.transactionRepository.AssertExpectations(t)
		tc.boundedBlockingQueue.AssertExpectations(t)
		tc.nonceProvider.AssertExpectations(t)
		tc.custodyProvider.AssertExpectations(t)
		tc.ethClientRegistry.AssertNotCalled(t, "GetClientForNetworkId", mock.Anything, mock.Anything)
	})

	t.Run("GasLimit override", func(t *testing.T) {
		tx := factory.CreateEntity(func(t *evmtransaction.EvmTransaction) {
			t.TransactionType = portcommon.TransactionTypeLegacy
		})
		custodyKey := custodyKeyFactory.CreateEntity(func(key *custodykey.CustodyKey) {
			key.DltAccountId = tx.FromAddress
		})
		tc := newTestContext()
		ctx := context.Background()
		overrideGasLimit, _ := amount.NewFromString("30000")
		cmd := &Command{
			TransactionHash: tx.TxId,
			GasLimit:        overrideGasLimit,
		}

		tc.boundedBlockingQueue.On("Put", ctx, validNetworkId, mock.Anything).Return(nil)

		newNonce, _ := amount.NewFromString("10")
		tc.nonceProvider.On("GetNonce", ctx, &nonceprovider.GetNonceRequest{
			NetworkId:    tx.NetworkId,
			DltAccountId: tx.FromAddress,
		}).Return(newNonce, nil)

		expectFindByTxId(tc, ctx, tx)
		expectCustodyKeyExists(tc, ctx, tx.FromAddress, custodyKey)
		expectSign(tc, ctx, buildExpectedSignRequest(
			tx,
			custodyKey,
			typedChainId,
			newNonce,
			overrideGasLimit,
			tx.GasPrice,
			tx.MaxPriorityFeePerGas,
			tx.MaxFeePerGas,
		))
		sender := registerSender(tc, ctx, tx)
		expectSave(tc, ctx, func(saved *evmtransaction.EvmTransaction) bool {
			return saved != nil &&
				saved.TxId == validTxId &&
				saved.GasLimit != nil &&
				saved.GasLimit.Equal(*overrideGasLimit)
		})
		expectPublish(tc, ctx, cmd.TransactionHash)

		resp, err := tc.handler.Handle(ctx, cmd)

		assert.Nil(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, validTxId, resp.NewTxId)
		sender.AssertExpectations(t)
		tc.transactionRepository.AssertExpectations(t)
		tc.boundedBlockingQueue.AssertExpectations(t)
		tc.nonceProvider.AssertExpectations(t)
		tc.custodyProvider.AssertExpectations(t)
		tc.ethClientRegistry.AssertNotCalled(t, "GetClientForNetworkId", mock.Anything, mock.Anything)
	})

	t.Run("useMultiplier=true", func(t *testing.T) {
		tx := factory.CreateEntity(func(t *evmtransaction.EvmTransaction) {
			t.TransactionType = portcommon.TransactionTypeLegacy
		})
		custodyKey := custodyKeyFactory.CreateEntity(func(key *custodykey.CustodyKey) {
			key.DltAccountId = tx.FromAddress
		})
		tc := newTestContext()
		ctx := context.Background()
		cmd := &Command{
			TransactionHash: tx.TxId,
			UseMultiplier:   true,
		}

		tc.boundedBlockingQueue.On("Put", ctx, validNetworkId, mock.Anything).Return(nil)

		network, _ := config.DltIngressConfig.DltIngress.GetDltIngressNetwork(tx.NetworkId)
		multipliedGasPrice := tx.GasPrice.MulWithDecimals(network.FeeMultiplier, 0)

		newNonce, _ := amount.NewFromString("10")
		tc.nonceProvider.On("GetNonce", ctx, &nonceprovider.GetNonceRequest{
			NetworkId:    tx.NetworkId,
			DltAccountId: tx.FromAddress,
		}).Return(newNonce, nil)

		expectFindByTxId(tc, ctx, tx)
		expectCustodyKeyExists(tc, ctx, tx.FromAddress, custodyKey)
		expectSign(tc, ctx, buildExpectedSignRequest(
			tx,
			custodyKey,
			typedChainId,
			newNonce,
			tx.GasLimit,
			&multipliedGasPrice,
			tx.MaxPriorityFeePerGas,
			tx.MaxFeePerGas,
		))
		sender := registerSender(tc, ctx, tx)
		expectSave(tc, ctx, func(saved *evmtransaction.EvmTransaction) bool {
			return saved != nil &&
				saved.TxId == validTxId &&
				saved.GasPrice != nil &&
				saved.GasPrice.Equal(multipliedGasPrice)
		})
		expectPublish(tc, ctx, cmd.TransactionHash)

		resp, err := tc.handler.Handle(ctx, cmd)

		assert.Nil(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, validTxId, resp.NewTxId)
		sender.AssertExpectations(t)
		tc.transactionRepository.AssertExpectations(t)
		tc.boundedBlockingQueue.AssertExpectations(t)
		tc.nonceProvider.AssertExpectations(t)
		tc.custodyProvider.AssertExpectations(t)
		tc.ethClientRegistry.AssertNotCalled(t, "GetClientForNetworkId", mock.Anything, mock.Anything)
	})

	t.Run("Nonce override", func(t *testing.T) {
		tx := factory.CreateEntity(func(t *evmtransaction.EvmTransaction) {
			t.TransactionType = portcommon.TransactionTypeLegacy
		})
		custodyKey := custodyKeyFactory.CreateEntity(func(key *custodykey.CustodyKey) {
			key.DltAccountId = tx.FromAddress
		})
		tc := newTestContext()
		ctx := context.Background()
		overrideNonce, _ := amount.NewFromString("99")
		cmd := &Command{
			TransactionHash: tx.TxId,
			Nonce:           overrideNonce,
		}

		tc.boundedBlockingQueue.On("Put", ctx, validNetworkId, mock.Anything).Return(nil)

		expectFindByTxId(tc, ctx, tx)
		expectCustodyKeyExists(tc, ctx, tx.FromAddress, custodyKey)
		expectSign(tc, ctx, buildExpectedSignRequest(
			tx,
			custodyKey,
			typedChainId,
			overrideNonce,
			tx.GasLimit,
			tx.GasPrice,
			tx.MaxPriorityFeePerGas,
			tx.MaxFeePerGas,
		))
		sender := registerSender(tc, ctx, tx)
		expectSave(tc, ctx, func(saved *evmtransaction.EvmTransaction) bool {
			return saved != nil &&
				saved.TxId == validTxId &&
				saved.Nonce != nil &&
				saved.Nonce.Equal(*overrideNonce)
		})
		expectPublish(tc, ctx, cmd.TransactionHash)

		resp, err := tc.handler.Handle(ctx, cmd)

		assert.Nil(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, validTxId, resp.NewTxId)
		sender.AssertExpectations(t)
		tc.transactionRepository.AssertExpectations(t)
		tc.boundedBlockingQueue.AssertExpectations(t)
		tc.nonceProvider.AssertNotCalled(t, "GetRemoteNonce", mock.Anything, mock.Anything, mock.Anything)
		tc.custodyProvider.AssertExpectations(t)
		tc.ethClientRegistry.AssertNotCalled(t, "GetClientForNetworkId", mock.Anything, mock.Anything)
	})

	t.Run("With OriginalTxId in tx", func(t *testing.T) {
		tx := factory.CreateEntity(func(t *evmtransaction.EvmTransaction) {
			t.TransactionType = portcommon.TransactionTypeLegacy
			t.OriginalTxId = new("test original transaction id")
		})
		custodyKey := custodyKeyFactory.CreateEntity(func(key *custodykey.CustodyKey) {
			key.DltAccountId = tx.FromAddress
		})
		tc := newTestContext()
		ctx := context.Background()
		cmd := &Command{TransactionHash: tx.TxId}

		tc.boundedBlockingQueue.On("Put", ctx, validNetworkId, mock.Anything).Return(nil)

		newNonce, _ := amount.NewFromString("10")
		tc.nonceProvider.On("GetNonce", ctx, &nonceprovider.GetNonceRequest{
			NetworkId:    tx.NetworkId,
			DltAccountId: tx.FromAddress,
		}).Return(newNonce, nil)

		expectFindByTxId(tc, ctx, tx)
		expectCustodyKeyExists(tc, ctx, tx.FromAddress, custodyKey)
		expectSign(tc, ctx, buildExpectedSignRequest(
			tx,
			custodyKey,
			typedChainId,
			newNonce,
			tx.GasLimit,
			tx.GasPrice,
			tx.MaxPriorityFeePerGas,
			tx.MaxFeePerGas,
		))
		sender := registerSender(tc, ctx, tx)
		expectSave(tc, ctx, func(saved *evmtransaction.EvmTransaction) bool {
			return saved != nil &&
				saved.TxId == validTxId &&
				saved.OriginalTxId == tx.OriginalTxId
		})
		expectPublish(tc, ctx, cmd.TransactionHash)

		resp, err := tc.handler.Handle(ctx, cmd)

		assert.Nil(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, validTxId, resp.NewTxId)
		sender.AssertExpectations(t)
		tc.transactionRepository.AssertExpectations(t)
		tc.boundedBlockingQueue.AssertExpectations(t)
		tc.nonceProvider.AssertExpectations(t)
		tc.custodyProvider.AssertExpectations(t)
		tc.ethClientRegistry.AssertNotCalled(t, "GetClientForNetworkId", mock.Anything, mock.Anything)
	})
}

func TestDltIngressRetryTransactionCommandHandler_TxTypeDynamicFee_Handle_Success(t *testing.T) {
	factory := txfactory.NewEvmTransactionTestFactory()
	custodyKeyFactory := custodykeyfactory.NewCustodyKeyTestFactory()
	typedChainId, _ := amount.NewFromString(validChainId)

	t.Run("same MaxPriorityFeePerGas and GasLimit", func(t *testing.T) {
		tx := factory.CreateEntityDynamicFee()
		custodyKey := custodyKeyFactory.CreateEntity(func(key *custodykey.CustodyKey) {
			key.DltAccountId = tx.FromAddress
		})
		tc := newTestContext()
		ctx := context.Background()
		cmd := &Command{TransactionHash: tx.TxId}

		tc.boundedBlockingQueue.On("Put", ctx, validNetworkId, mock.Anything).Return(nil)

		newNonce, _ := amount.NewFromString("10")
		tc.nonceProvider.On("GetNonce", ctx, &nonceprovider.GetNonceRequest{
			NetworkId:    tx.NetworkId,
			DltAccountId: tx.FromAddress,
		}).Return(newNonce, nil)

		expectFindByTxId(tc, ctx, tx)
		expectCustodyKeyExists(tc, ctx, tx.FromAddress, custodyKey)
		expectedMaxFeePerGas := expectCalculateMaxFee(tc, ctx, tx, tx.MaxPriorityFeePerGas)
		expectSign(tc, ctx, buildExpectedSignRequest(
			tx,
			custodyKey,
			typedChainId,
			newNonce,
			tx.GasLimit,
			tx.GasPrice,
			tx.MaxPriorityFeePerGas,
			&expectedMaxFeePerGas,
		))
		sender := registerSender(tc, ctx, tx)
		expectSave(tc, ctx, func(saved *evmtransaction.EvmTransaction) bool {
			return saved != nil &&
				saved.TxId == validTxId &&
				saved.MaxPriorityFeePerGas != nil &&
				saved.MaxPriorityFeePerGas.Equal(*tx.MaxPriorityFeePerGas) &&
				saved.MaxFeePerGas != nil &&
				saved.MaxFeePerGas.Equal(expectedMaxFeePerGas)
		})
		expectPublish(tc, ctx, cmd.TransactionHash)

		resp, err := tc.handler.Handle(ctx, cmd)

		assert.Nil(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, validTxId, resp.NewTxId)
		sender.AssertExpectations(t)
		tc.transactionRepository.AssertExpectations(t)
		tc.boundedBlockingQueue.AssertExpectations(t)
		tc.nonceProvider.AssertExpectations(t)
		tc.custodyProvider.AssertExpectations(t)
		tc.ethClientRegistry.AssertExpectations(t)
		tc.ethClient.AssertExpectations(t)
	})

	t.Run("MaxPriorityFeePerGas override", func(t *testing.T) {
		tx := factory.CreateEntityDynamicFee()
		custodyKey := custodyKeyFactory.CreateEntity(func(key *custodykey.CustodyKey) {
			key.DltAccountId = tx.FromAddress
		})
		tc := newTestContext()
		ctx := context.Background()
		overridePriorityFee, _ := amount.NewFromString("3000000000")
		cmd := &Command{
			TransactionHash:      tx.TxId,
			MaxPriorityFeePerGas: overridePriorityFee,
		}

		tc.boundedBlockingQueue.On("Put", ctx, validNetworkId, mock.Anything).Return(nil)

		newNonce, _ := amount.NewFromString("10")
		tc.nonceProvider.On("GetNonce", ctx, &nonceprovider.GetNonceRequest{
			NetworkId:    tx.NetworkId,
			DltAccountId: tx.FromAddress,
		}).Return(newNonce, nil)

		expectFindByTxId(tc, ctx, tx)
		expectCustodyKeyExists(tc, ctx, tx.FromAddress, custodyKey)
		expectedMaxFeePerGas := expectCalculateMaxFee(tc, ctx, tx, overridePriorityFee)
		expectSign(tc, ctx, buildExpectedSignRequest(
			tx,
			custodyKey,
			typedChainId,
			newNonce,
			tx.GasLimit,
			tx.GasPrice,
			overridePriorityFee,
			&expectedMaxFeePerGas,
		))
		sender := registerSender(tc, ctx, tx)
		expectSave(tc, ctx, func(saved *evmtransaction.EvmTransaction) bool {
			return saved != nil &&
				saved.TxId == validTxId &&
				saved.MaxPriorityFeePerGas != nil &&
				saved.MaxPriorityFeePerGas.Equal(*overridePriorityFee) &&
				saved.MaxFeePerGas != nil &&
				saved.MaxFeePerGas.Equal(expectedMaxFeePerGas)
		})
		expectPublish(tc, ctx, cmd.TransactionHash)

		resp, err := tc.handler.Handle(ctx, cmd)

		assert.Nil(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, validTxId, resp.NewTxId)
		sender.AssertExpectations(t)
		tc.transactionRepository.AssertExpectations(t)
		tc.boundedBlockingQueue.AssertExpectations(t)
		tc.nonceProvider.AssertExpectations(t)
		tc.custodyProvider.AssertExpectations(t)
		tc.ethClientRegistry.AssertExpectations(t)
		tc.ethClient.AssertExpectations(t)
	})

	t.Run("GasLimit override", func(t *testing.T) {
		tx := factory.CreateEntityDynamicFee()
		custodyKey := custodyKeyFactory.CreateEntity(func(key *custodykey.CustodyKey) {
			key.DltAccountId = tx.FromAddress
		})
		tc := newTestContext()
		ctx := context.Background()
		overrideGasLimit, _ := amount.NewFromString("3000000")
		cmd := &Command{
			TransactionHash: tx.TxId,
			GasLimit:        overrideGasLimit,
		}

		tc.boundedBlockingQueue.On("Put", ctx, validNetworkId, mock.Anything).Return(nil)

		newNonce, _ := amount.NewFromString("10")
		tc.nonceProvider.On("GetNonce", ctx, &nonceprovider.GetNonceRequest{
			NetworkId:    tx.NetworkId,
			DltAccountId: tx.FromAddress,
		}).Return(newNonce, nil)

		expectFindByTxId(tc, ctx, tx)
		expectCustodyKeyExists(tc, ctx, tx.FromAddress, custodyKey)
		expectedMaxFeePerGas := expectCalculateMaxFee(tc, ctx, tx, tx.MaxPriorityFeePerGas)
		expectSign(tc, ctx, buildExpectedSignRequest(
			tx,
			custodyKey,
			typedChainId,
			newNonce,
			overrideGasLimit,
			tx.GasPrice,
			tx.MaxPriorityFeePerGas,
			&expectedMaxFeePerGas,
		))
		sender := registerSender(tc, ctx, tx)
		expectSave(tc, ctx, func(saved *evmtransaction.EvmTransaction) bool {
			return saved != nil &&
				saved.TxId == validTxId &&
				saved.GasLimit != nil &&
				saved.GasLimit.Equal(*overrideGasLimit)
		})
		expectPublish(tc, ctx, cmd.TransactionHash)

		resp, err := tc.handler.Handle(ctx, cmd)

		assert.Nil(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, validTxId, resp.NewTxId)
		sender.AssertExpectations(t)
		tc.transactionRepository.AssertExpectations(t)
		tc.boundedBlockingQueue.AssertExpectations(t)
		tc.nonceProvider.AssertExpectations(t)
		tc.custodyProvider.AssertExpectations(t)
		tc.ethClientRegistry.AssertExpectations(t)
		tc.ethClient.AssertExpectations(t)
	})

	t.Run("useMultiplier=true", func(t *testing.T) {
		tx := factory.CreateEntityDynamicFee()
		custodyKey := custodyKeyFactory.CreateEntity(func(key *custodykey.CustodyKey) {
			key.DltAccountId = tx.FromAddress
		})
		tc := newTestContext()
		ctx := context.Background()
		cmd := &Command{
			TransactionHash: tx.TxId,
			UseMultiplier:   true,
		}

		network, _ := config.DltIngressConfig.DltIngress.GetDltIngressNetwork(tx.NetworkId)
		multipliedPriorityFee := tx.MaxPriorityFeePerGas.MulWithDecimals(network.FeeMultiplier, 0)

		tc.boundedBlockingQueue.On("Put", ctx, validNetworkId, mock.Anything).Return(nil)

		newNonce, _ := amount.NewFromString("10")
		tc.nonceProvider.On("GetNonce", ctx, &nonceprovider.GetNonceRequest{
			NetworkId:    tx.NetworkId,
			DltAccountId: tx.FromAddress,
		}).Return(newNonce, nil)

		expectFindByTxId(tc, ctx, tx)
		expectCustodyKeyExists(tc, ctx, tx.FromAddress, custodyKey)
		expectedMaxFeePerGas := expectCalculateMaxFee(tc, ctx, tx, &multipliedPriorityFee)
		expectSign(tc, ctx, buildExpectedSignRequest(
			tx,
			custodyKey,
			typedChainId,
			newNonce,
			tx.GasLimit,
			tx.GasPrice,
			&multipliedPriorityFee,
			&expectedMaxFeePerGas,
		))
		sender := registerSender(tc, ctx, tx)
		expectSave(tc, ctx, func(saved *evmtransaction.EvmTransaction) bool {
			return saved != nil &&
				saved.TxId == validTxId &&
				saved.MaxPriorityFeePerGas != nil &&
				saved.MaxPriorityFeePerGas.Equal(multipliedPriorityFee) &&
				saved.MaxFeePerGas != nil &&
				saved.MaxFeePerGas.Equal(expectedMaxFeePerGas)
		})
		expectPublish(tc, ctx, cmd.TransactionHash)

		resp, err := tc.handler.Handle(ctx, cmd)

		assert.Nil(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, validTxId, resp.NewTxId)
		sender.AssertExpectations(t)
		tc.transactionRepository.AssertExpectations(t)
		tc.boundedBlockingQueue.AssertExpectations(t)
		tc.nonceProvider.AssertExpectations(t)
		tc.custodyProvider.AssertExpectations(t)
		tc.ethClientRegistry.AssertExpectations(t)
		tc.ethClient.AssertExpectations(t)
	})

	t.Run("Nonce override", func(t *testing.T) {
		tx := factory.CreateEntityDynamicFee()
		custodyKey := custodyKeyFactory.CreateEntity(func(key *custodykey.CustodyKey) {
			key.DltAccountId = tx.FromAddress
		})
		tc := newTestContext()
		ctx := context.Background()
		overrideNonce, _ := amount.NewFromString("99")
		cmd := &Command{
			TransactionHash: tx.TxId,
			Nonce:           overrideNonce,
		}

		tc.boundedBlockingQueue.On("Put", ctx, validNetworkId, mock.Anything).Return(nil)

		expectFindByTxId(tc, ctx, tx)
		expectCustodyKeyExists(tc, ctx, tx.FromAddress, custodyKey)
		expectedMaxFeePerGas := expectCalculateMaxFee(tc, ctx, tx, tx.MaxPriorityFeePerGas)
		expectSign(tc, ctx, buildExpectedSignRequest(
			tx,
			custodyKey,
			typedChainId,
			overrideNonce,
			tx.GasLimit,
			tx.GasPrice,
			tx.MaxPriorityFeePerGas,
			&expectedMaxFeePerGas,
		))
		sender := registerSender(tc, ctx, tx)
		expectSave(tc, ctx, func(saved *evmtransaction.EvmTransaction) bool {
			return saved != nil &&
				saved.TxId == validTxId &&
				saved.Nonce != nil &&
				saved.Nonce.Equal(*overrideNonce)
		})
		expectPublish(tc, ctx, cmd.TransactionHash)

		resp, err := tc.handler.Handle(ctx, cmd)

		assert.Nil(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, validTxId, resp.NewTxId)
		sender.AssertExpectations(t)
		tc.transactionRepository.AssertExpectations(t)
		tc.boundedBlockingQueue.AssertExpectations(t)
		tc.nonceProvider.AssertNotCalled(t, "GetRemoteNonce", mock.Anything, mock.Anything, mock.Anything)
		tc.custodyProvider.AssertExpectations(t)
		tc.ethClientRegistry.AssertExpectations(t)
		tc.ethClient.AssertExpectations(t)
	})
}

func TestDltIngressRetryTransactionCommandHandler_Handle_TxNotFound(t *testing.T) {
	tc := newTestContext()
	dummyTxId := "dummyTxId"

	ctx := context.Background()
	cmd := &Command{
		TransactionHash: dummyTxId,
	}

	tc.transactionRepository.On("FindByTxId", ctx, dummyTxId).Return(nil, domainerrors.NewEntityNotFoundDomainError("Transaction", uuid.Nil))

	resp, err := tc.handler.Handle(ctx, cmd)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, coreerror.ErrNotFound)
}

func TestDltIngressRetryTransactionCommandHandler_Handle_NetworkNotFound(t *testing.T) {
	factory := txfactory.NewEvmTransactionTestFactory()
	tx := factory.CreateEntityDynamicFee(func(t *evmtransaction.EvmTransaction) {
		t.NetworkId = "dummyNetwork"
	})

	tc := newTestContext()

	ctx := context.Background()
	cmd := &Command{
		TransactionHash: tx.TxId,
	}

	tc.transactionRepository.On("FindByTxId", ctx, tx.TxId).Return(tx, nil)

	resp, err := tc.handler.Handle(ctx, cmd)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, "Entity Network with dummyNetwork not found", err.Error())
}

func TestDltIngressRetryTransactionCommandHandler_Handle_GasLimitAboveCeiling_PassesThrough(t *testing.T) {
	factory := txfactory.NewEvmTransactionTestFactory()
	tx := factory.CreateEntity(func(t *evmtransaction.EvmTransaction) {
		t.TransactionType = portcommon.TransactionTypeLegacy
	})
	custodyKeyFactory := custodykeyfactory.NewCustodyKeyTestFactory()
	custodyKey := custodyKeyFactory.CreateEntity(func(key *custodykey.CustodyKey) {
		key.DltAccountId = tx.FromAddress
	})

	tc := newTestContext()
	ctx := context.Background()

	// Twenty times the configured ceiling. This is an admin override, so it is applied as given:
	// the caller owns the consequences and the relay is what rejects it on send.
	overrideGasLimit, _ := amount.NewFromString("300000000")
	cmd := &Command{
		TransactionHash: tx.TxId,
		GasLimit:        overrideGasLimit,
	}

	tc.boundedBlockingQueue.On("Put", ctx, validNetworkId, mock.Anything).Return(nil)

	newNonce, _ := amount.NewFromString("10")
	tc.nonceProvider.On("GetNonce", ctx, &nonceprovider.GetNonceRequest{
		NetworkId:    tx.NetworkId,
		DltAccountId: tx.FromAddress,
	}).Return(newNonce, nil)

	typedChainId, _ := amount.NewFromString(validChainId)
	expectFindByTxId(tc, ctx, tx)
	expectCustodyKeyExists(tc, ctx, tx.FromAddress, custodyKey)
	expectSign(tc, ctx, buildExpectedSignRequest(
		tx,
		custodyKey,
		typedChainId,
		newNonce,
		overrideGasLimit,
		tx.GasPrice,
		tx.MaxPriorityFeePerGas,
		tx.MaxFeePerGas,
	))
	sender := registerSender(tc, ctx, tx)
	expectSave(tc, ctx, func(saved *evmtransaction.EvmTransaction) bool {
		return saved != nil && saved.GasLimit != nil && saved.GasLimit.Equal(*overrideGasLimit)
	})
	expectPublish(tc, ctx, cmd.TransactionHash)

	resp, err := tc.handler.Handle(ctx, cmd)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, validTxId, resp.NewTxId)
	sender.AssertExpectations(t)
	tc.custodyProvider.AssertExpectations(t)
	tc.transactionRepository.AssertExpectations(t)
}

func TestDltIngressRetryTransactionCommandHandler_Handle_SendFailure_HardDeletesPersistedTxAndReleasesQueueSlot(t *testing.T) {
	factory := txfactory.NewEvmTransactionTestFactory()
	tx := factory.CreateEntity(func(t *evmtransaction.EvmTransaction) {
		t.TransactionType = portcommon.TransactionTypeLegacy
	})
	custodyKeyFactory := custodykeyfactory.NewCustodyKeyTestFactory()
	custodyKey := custodyKeyFactory.CreateEntity(func(key *custodykey.CustodyKey) {
		key.DltAccountId = tx.FromAddress
	})

	tc := newTestContext()
	ctx := context.Background()
	cmd := &Command{TransactionHash: tx.TxId}

	tc.boundedBlockingQueue.On("Put", ctx, validNetworkId, mock.Anything).Return(nil)
	tc.boundedBlockingQueue.On("Take", ctx, validNetworkId).Return(nil)

	newNonce, _ := amount.NewFromString("10")
	tc.nonceProvider.On("GetNonce", ctx, &nonceprovider.GetNonceRequest{
		NetworkId:    tx.NetworkId,
		DltAccountId: tx.FromAddress,
	}).Return(newNonce, nil)

	typedChainId, _ := amount.NewFromString(validChainId)
	expectFindByTxId(tc, ctx, tx)
	expectCustodyKeyExists(tc, ctx, tx.FromAddress, custodyKey)
	expectSign(tc, ctx, buildExpectedSignRequest(
		tx,
		custodyKey,
		typedChainId,
		newNonce,
		tx.GasLimit,
		tx.GasPrice,
		tx.MaxPriorityFeePerGas,
		tx.MaxFeePerGas,
	))

	var savedTx *evmtransaction.EvmTransaction
	expectSave(tc, ctx, func(saved *evmtransaction.EvmTransaction) bool {
		savedTx = saved
		return true
	})
	tc.transactionRepository.
		On("HardDelete", ctx, mock.MatchedBy(func(deleted *evmtransaction.EvmTransaction) bool {
			return deleted == savedTx
		})).
		Return(nil)

	transactionSender := new(tsmock.TransactionSenderMock)
	tc.transactionSenderRegistry.Register(common.EVM, transactionSender)
	transactionSender.On("SendTransaction", ctx, mock.Anything).Return(nil, assert.AnError)

	resp, err := tc.handler.Handle(ctx, cmd)

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, assert.AnError)
	tc.boundedBlockingQueue.AssertExpectations(t)
	tc.transactionRepository.AssertExpectations(t)
	tc.transactionRepository.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
	tc.eventBus.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
}
