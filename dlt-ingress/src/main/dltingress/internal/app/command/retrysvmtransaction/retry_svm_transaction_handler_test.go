package retrysvmtransaction

import (
	"context"
	"testing"

	custodykeyfactory "dlt-ingress/src/main/dltingress/internal/domain/custodykey"
	servicemock "dlt-ingress/src/main/dltingress/internal/domain/custodykey/service/mock"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction/svmtransaction"
	"dlt-ingress/src/main/dltingress/internal/infra/contracttransactionbuilder"
	mocksvmtransactionrepo "dlt-ingress/src/main/dltingress/internal/infra/repository/transaction/svmtransaction/mock"
	"dlt-ingress/src/main/dltingress/internal/infra/svm"
	tsmock "dlt-ingress/src/main/dltingress/internal/infra/transanctionsender/mock"

	bbqmock "dlt-ingress/src/main/dltingress/internal/infra/boundedblockingqueue/mock"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"dlt-ingress/src/main/dltingress/internal/domain/custodykey"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction"
	"dlt-ingress/src/main/dltingress/internal/infra/custody"
	custodymock "dlt-ingress/src/main/dltingress/internal/infra/custody/mock"
	"dlt-ingress/src/main/dltingress/internal/infra/transanctionsender"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"
	testbus "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/test/core/event"
)

const (
	validNetworkId     = "svm-1"
	validDlt           = "SVM"
	validUrl           = "https://api.mainnet-beta.solana.com"
	validMaxTxPoolSize = 100
	validTxId          = "5s9f1F8y8n2q1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6"

	oldBlockhash = "11111111111111111111111111111111"
	newBlockhash = "So11111111111111111111111111111111111111112"

	testIdlJSON = `{
  "address": "11111111111111111111111111111112",
  "metadata": { "name": "sample", "version": "0.1.0", "spec": "0.1.0" },
  "instructions": [
    {
      "name": "transfer",
      "discriminator": [1, 2, 3, 4, 5, 6, 7, 8],
      "accounts": [
        { "name": "user",          "writable": true, "signer": true },
        { "name": "recipient",     "writable": true },
        { "name": "systemProgram", "address": "11111111111111111111111111111111" }
      ],
      "args": [
        { "name": "amount", "type": "u64" }
      ]
    }
  ]
}`

	// testMultiSignerIdlJSON mirrors instructions like deploy_mint, where a fresh account
	// being created (newAccount) must sign alongside the fee payer (payer).
	testMultiSignerIdlJSON = `{
  "address": "11111111111111111111111111111112",
  "metadata": { "name": "sample", "version": "0.1.0", "spec": "0.1.0" },
  "instructions": [
    {
      "name": "create",
      "discriminator": [1, 2, 3, 4, 5, 6, 7, 8],
      "accounts": [
        { "name": "payer",         "writable": true, "signer": true },
        { "name": "newAccount",    "writable": true, "signer": true },
        { "name": "systemProgram", "address": "11111111111111111111111111111111" }
      ],
      "args": [
        { "name": "amount", "type": "u64" }
      ]
    }
  ]
}`
)

var (
	testFeePayer  = solana.MustPublicKeyFromBase58("So11111111111111111111111111111111111111112")
	testRecipient = solana.MustPublicKeyFromBase58("4Nd1mYQbHmZ2HhZJ4uEjvR5tQv3qHN3F8XJyR9KkPp2C")
)

func init() {
	typedFeeMultiplier, _ := amount.NewFromString("1.2")
	typedMaxCuLimit, _ := amount.NewFromString("1400000")
	typedMaxCuPrice, _ := amount.NewFromString("1000000")

	config.DltIngressConfig = &config.Config{}
	config.DltIngressConfig.DltIngress.Networks = []*config.NetworkConfig{
		{
			Id:            validNetworkId,
			Url:           validUrl,
			Dlt:           validDlt,
			FeeMultiplier: *typedFeeMultiplier,
			MaxCuLimit:    *typedMaxCuLimit,
			MaxCuPrice:    *typedMaxCuPrice,
			MaxTxPoolSize: validMaxTxPoolSize,
		},
	}
}

type testContext struct {
	handler                        *CommandHandler
	eventBus                       *testbus.EventBusMock
	custodyProvider                *custodymock.CustodyProviderMock
	svmClientRegistry              *svm.SvmClientRegistryMock
	svmClient                      *svm.SvmClientMock
	transactionSenderRegistry      *transanctionsender.Registry
	transactionRepository          *mocksvmtransactionrepo.Postgres
	boundedBlockingQueue           *bbqmock.BoundedBlockingQueueMock
	custodyKeyExistMultipleService *servicemock.CustodyKeyExistMultipleMock
}

func newTestContext() *testContext {
	eventBus := new(testbus.EventBusMock)
	svmClientRegistry := new(svm.SvmClientRegistryMock)
	svmClient := new(svm.SvmClientMock)
	custodyProvider := new(custodymock.CustodyProviderMock)
	transactionSenderRegistry := transanctionsender.NewRegistry()
	transactionRepository := new(mocksvmtransactionrepo.Postgres)
	bbq := new(bbqmock.BoundedBlockingQueueMock)
	custodyKeyExistMultipleService := new(servicemock.CustodyKeyExistMultipleMock)

	handler := NewCommandHandler(
		eventBus,
		custodyProvider,
		svmClientRegistry,
		*transactionSenderRegistry,
		transactionRepository,
		bbq,
		custodyKeyExistMultipleService,
	)

	return &testContext{
		handler:                        handler,
		eventBus:                       eventBus,
		custodyProvider:                custodyProvider,
		svmClientRegistry:              svmClientRegistry,
		svmClient:                      svmClient,
		transactionSenderRegistry:      transactionSenderRegistry,
		transactionRepository:          transactionRepository,
		boundedBlockingQueue:           bbq,
		custodyKeyExistMultipleService: custodyKeyExistMultipleService,
	}
}

// buildOldTx builds a realistic persisted SvmTransaction fixture by running it through the real
// SVM IDL builder, the same way the original send path would have produced it.
func buildOldTx(t *testing.T, cuLimit, cuPrice *amount.Amount) *svmtransaction.SvmTransaction {
	t.Helper()
	builder, err := contracttransactionbuilder.NewSvmIdlTransactionBuilder([]byte(testIdlJSON))
	require.NoError(t, err)

	req := contracttransactionbuilder.NewSVMBuildTransactionRequest(
		testFeePayer.String(),
		"11111111111111111111111111111112",
		"transfer",
		map[string]any{"recipient": testRecipient.String(), "amount": uint64(1)},
		oldBlockhash,
		cuPrice,
		cuLimit,
	)
	resp, err := builder.BuildTransaction(req)
	require.NoError(t, err)

	tx, err := svmtransaction.NewSvmTransaction(
		"old-tx-signature",
		validNetworkId,
		validUrl,
		resp.SVMTransactionResponse.FeePayer,
		resp.SVMTransactionResponse.RecentBlockhash,
		resp.SVMTransactionResponse.SerializedTransaction,
		cuLimit,
		cuPrice,
	)
	require.NoError(t, err)
	return tx
}

func expectFreshBlockhash(tc *testContext, ctx context.Context) {
	tc.svmClientRegistry.On("GetClientForNetworkId", validNetworkId).Return(tc.svmClient, nil)
	tc.svmClient.On("GetRecentBlockhash", ctx).Return(newBlockhash, nil)
}

func expectFindByTxId(tc *testContext, ctx context.Context, tx *svmtransaction.SvmTransaction) {
	tc.transactionRepository.On("FindByTxId", ctx, tx.TxId).Return(tx, nil)
}

func expectCustodyKeysExist(tc *testContext, ctx context.Context, signerIds []string, custodyKeys []*custodykey.CustodyKey) {
	tc.custodyKeyExistMultipleService.On("Execute", ctx, signerIds).Return(custodyKeys, nil)
}

// expectSingleSignerCustodyKey is a convenience for the common single-signer (fee payer only)
// fixtures used throughout this file.
func expectSingleSignerCustodyKey(tc *testContext, ctx context.Context, feePayer string, custodyKey *custodykey.CustodyKey) {
	expectCustodyKeysExist(tc, ctx, []string{feePayer}, []*custodykey.CustodyKey{custodyKey})
}

func expectPublish(tc *testContext, ctx context.Context, oldTxId string) {
	tc.eventBus.
		On("Publish", ctx, mock.MatchedBy(func(evt transaction.RetriedEvent) bool {
			return evt.NewTxId == validTxId && evt.OldTxId == oldTxId
		})).
		Return(nil)
}

func registerSender(tc *testContext, ctx context.Context) *tsmock.TransactionSenderMock {
	transactionSender := new(tsmock.TransactionSenderMock)
	tc.transactionSenderRegistry.Register(common.SVM, transactionSender)

	transactionSender.
		On("SendTransaction", ctx, mock.MatchedBy(func(req transanctionsender.SendTransactionRequest) bool {
			return req.Dlt == validDlt && req.NetworkId == validNetworkId
		})).
		Return(&transanctionsender.SendTransactionResponse{TxId: validTxId}, nil)

	return transactionSender
}

// decodedInstructionKinds decodes the signed request's transaction and returns, for each
// instruction in order, "cu-limit", "cu-price", or "other".
func decodedSignRequestTx(t *testing.T, req *custody.SignRequest) *solana.Transaction {
	t.Helper()
	require.NotNil(t, req.Transaction)
	require.NotNil(t, req.Transaction.SVMTransactionResponse)
	tx, err := solana.TransactionFromBase64(req.Transaction.SVMTransactionResponse.SerializedTransaction)
	require.NoError(t, err)
	return tx
}

func expectSign(tc *testContext, ctx context.Context, matcher func(req *custody.SignRequest) bool) {
	tc.custodyProvider.
		On("Sign", ctx, mock.MatchedBy(matcher)).
		Return(&custody.SignResponse{SignedTransaction: "signed-tx", TxId: validTxId}, nil)
}

func expectSave(tc *testContext, ctx context.Context, matcher func(saved *svmtransaction.SvmTransaction) bool) {
	tc.transactionRepository.
		On("Save", ctx, mock.MatchedBy(matcher)).
		Return(nil)
}

func TestDltIngressRetrySvmTransactionCommandHandler_Handle_Success_ReusesOriginalCuValues(t *testing.T) {
	cuLimit, _ := amount.NewFromString("200000")
	cuPrice, _ := amount.NewFromString("5000")
	oldTx := buildOldTx(t, cuLimit, cuPrice)

	custodyKeyFactory := custodykeyfactory.NewCustodyKeyTestFactory()
	custodyKey := custodyKeyFactory.CreateEntity(func(k *custodykey.CustodyKey) {
		k.DltAccountId = oldTx.FeePayer
	})

	tc := newTestContext()
	ctx := context.Background()
	cmd := &Command{TransactionHash: oldTx.TxId}

	tc.boundedBlockingQueue.On("Put", ctx, validNetworkId, mock.Anything).Return(nil)
	expectFindByTxId(tc, ctx, oldTx)
	expectSingleSignerCustodyKey(tc, ctx, oldTx.FeePayer, custodyKey)
	expectFreshBlockhash(tc, ctx)

	expectSign(tc, ctx, func(req *custody.SignRequest) bool {
		if req.Dlt != validDlt || len(req.CustodyKeys) != 1 || req.CustodyKeys[0] != custodyKey {
			return false
		}
		tx := decodedSignRequestTx(t, req)
		if tx.Message.RecentBlockhash.String() != newBlockhash {
			return false
		}
		require.Len(t, tx.Message.Instructions, 3)
		assert.Equal(t, computebudget.Instruction_SetComputeUnitLimit, []byte(tx.Message.Instructions[0].Data)[0])
		assert.Equal(t, computebudget.Instruction_SetComputeUnitPrice, []byte(tx.Message.Instructions[1].Data)[0])
		return req.Transaction.SVMTransactionResponse.CuLimit.Equal(*cuLimit) &&
			req.Transaction.SVMTransactionResponse.CuPrice.Equal(*cuPrice)
	})

	sender := registerSender(tc, ctx)
	expectSave(tc, ctx, func(saved *svmtransaction.SvmTransaction) bool {
		return saved != nil &&
			saved.TxId == validTxId &&
			saved.OriginalTxId != nil && *saved.OriginalTxId == oldTx.TxId &&
			saved.CuLimit.Equal(*cuLimit) &&
			saved.CuPrice.Equal(*cuPrice)
	})
	expectPublish(tc, ctx, cmd.TransactionHash)

	resp, err := tc.handler.Handle(ctx, cmd)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, validTxId, resp.NewTxId)
	sender.AssertExpectations(t)
	tc.transactionRepository.AssertExpectations(t)
	tc.boundedBlockingQueue.AssertExpectations(t)
	tc.custodyProvider.AssertExpectations(t)
	tc.svmClientRegistry.AssertExpectations(t)
	tc.svmClient.AssertExpectations(t)
}

func TestDltIngressRetrySvmTransactionCommandHandler_Handle_CuLimitAndCuPriceOverride(t *testing.T) {
	originalCuLimit, _ := amount.NewFromString("200000")
	originalCuPrice, _ := amount.NewFromString("5000")
	oldTx := buildOldTx(t, originalCuLimit, originalCuPrice)

	custodyKeyFactory := custodykeyfactory.NewCustodyKeyTestFactory()
	custodyKey := custodyKeyFactory.CreateEntity(func(k *custodykey.CustodyKey) {
		k.DltAccountId = oldTx.FeePayer
	})

	tc := newTestContext()
	ctx := context.Background()
	overrideCuLimit, _ := amount.NewFromString("300000")
	overrideCuPrice, _ := amount.NewFromString("9000")
	cmd := &Command{
		TransactionHash: oldTx.TxId,
		CuLimit:         overrideCuLimit,
		CuPrice:         overrideCuPrice,
	}

	tc.boundedBlockingQueue.On("Put", ctx, validNetworkId, mock.Anything).Return(nil)
	expectFindByTxId(tc, ctx, oldTx)
	expectSingleSignerCustodyKey(tc, ctx, oldTx.FeePayer, custodyKey)
	expectFreshBlockhash(tc, ctx)

	expectSign(tc, ctx, func(req *custody.SignRequest) bool {
		return req.Transaction.SVMTransactionResponse.CuLimit.Equal(*overrideCuLimit) &&
			req.Transaction.SVMTransactionResponse.CuPrice.Equal(*overrideCuPrice)
	})

	sender := registerSender(tc, ctx)
	expectSave(tc, ctx, func(saved *svmtransaction.SvmTransaction) bool {
		return saved.CuLimit.Equal(*overrideCuLimit) && saved.CuPrice.Equal(*overrideCuPrice)
	})
	expectPublish(tc, ctx, cmd.TransactionHash)

	resp, err := tc.handler.Handle(ctx, cmd)

	require.NoError(t, err)
	assert.Equal(t, validTxId, resp.NewTxId)
	sender.AssertExpectations(t)
	tc.custodyProvider.AssertExpectations(t)
	tc.transactionRepository.AssertExpectations(t)
}

// Regression test: a transaction can require signatures beyond the fee payer's — e.g. a fresh
// account being created, like deploy_mint's `mint` — and all of them must be resolved and
// re-signed on retry, or the rebuilt transaction still reserves a signer slot that never gets
// signed and fails signature verification on send.
func TestDltIngressRetrySvmTransactionCommandHandler_Handle_MultipleSigners_AllGetResigned(t *testing.T) {
	cuLimit, _ := amount.NewFromString("200000")
	cuPrice, _ := amount.NewFromString("5000")
	secondSigner := solana.NewWallet().PublicKey()

	builder, err := contracttransactionbuilder.NewSvmIdlTransactionBuilder([]byte(testMultiSignerIdlJSON))
	require.NoError(t, err)
	req := contracttransactionbuilder.NewSVMBuildTransactionRequest(
		testFeePayer.String(),
		"11111111111111111111111111111112",
		"create",
		map[string]any{"newAccount": secondSigner.String(), "amount": uint64(1)},
		oldBlockhash,
		cuPrice,
		cuLimit,
	)
	buildResp, err := builder.BuildTransaction(req)
	require.NoError(t, err)

	oldTx, err := svmtransaction.NewSvmTransaction(
		"old-tx-multi-signer",
		validNetworkId,
		validUrl,
		buildResp.SVMTransactionResponse.FeePayer,
		buildResp.SVMTransactionResponse.RecentBlockhash,
		buildResp.SVMTransactionResponse.SerializedTransaction,
		cuLimit,
		cuPrice,
	)
	require.NoError(t, err)

	custodyKeyFactory := custodykeyfactory.NewCustodyKeyTestFactory()
	feePayerKey := custodyKeyFactory.CreateEntity(func(k *custodykey.CustodyKey) { k.DltAccountId = testFeePayer.String() })
	secondSignerKey := custodyKeyFactory.CreateEntity(func(k *custodykey.CustodyKey) { k.DltAccountId = secondSigner.String() })

	tc := newTestContext()
	ctx := context.Background()
	cmd := &Command{TransactionHash: oldTx.TxId}

	tc.boundedBlockingQueue.On("Put", ctx, validNetworkId, mock.Anything).Return(nil)
	expectFindByTxId(tc, ctx, oldTx)
	tc.custodyKeyExistMultipleService.
		On("Execute", ctx, mock.MatchedBy(func(ids []string) bool {
			return len(ids) == 2 && containsAll(ids, testFeePayer.String(), secondSigner.String())
		})).
		Return([]*custodykey.CustodyKey{feePayerKey, secondSignerKey}, nil)
	expectFreshBlockhash(tc, ctx)

	expectSign(tc, ctx, func(req *custody.SignRequest) bool {
		return len(req.CustodyKeys) == 2 &&
			containsAll([]string{req.CustodyKeys[0].DltAccountId, req.CustodyKeys[1].DltAccountId},
				testFeePayer.String(), secondSigner.String())
	})

	sender := registerSender(tc, ctx)
	expectSave(tc, ctx, func(saved *svmtransaction.SvmTransaction) bool { return saved != nil })
	expectPublish(tc, ctx, cmd.TransactionHash)

	resp, err := tc.handler.Handle(ctx, cmd)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, validTxId, resp.NewTxId)
	sender.AssertExpectations(t)
	tc.custodyKeyExistMultipleService.AssertExpectations(t)
}

func containsAll(haystack []string, wanted ...string) bool {
	set := make(map[string]struct{}, len(haystack))
	for _, v := range haystack {
		set[v] = struct{}{}
	}
	for _, w := range wanted {
		if _, ok := set[w]; !ok {
			return false
		}
	}
	return true
}

func TestDltIngressRetrySvmTransactionCommandHandler_Handle_UseMultiplier_BumpsCuPrice(t *testing.T) {
	originalCuLimit, _ := amount.NewFromString("200000")
	originalCuPrice, _ := amount.NewFromString("5000")
	oldTx := buildOldTx(t, originalCuLimit, originalCuPrice)

	custodyKeyFactory := custodykeyfactory.NewCustodyKeyTestFactory()
	custodyKey := custodyKeyFactory.CreateEntity(func(k *custodykey.CustodyKey) {
		k.DltAccountId = oldTx.FeePayer
	})

	tc := newTestContext()
	ctx := context.Background()
	cmd := &Command{TransactionHash: oldTx.TxId, UseMultiplier: true}

	network, _ := config.DltIngressConfig.DltIngress.GetDltIngressNetwork(validNetworkId)
	multipliedCuPrice := originalCuPrice.MulWithDecimals(network.FeeMultiplier, 0)

	tc.boundedBlockingQueue.On("Put", ctx, validNetworkId, mock.Anything).Return(nil)
	expectFindByTxId(tc, ctx, oldTx)
	expectSingleSignerCustodyKey(tc, ctx, oldTx.FeePayer, custodyKey)
	expectFreshBlockhash(tc, ctx)

	expectSign(tc, ctx, func(req *custody.SignRequest) bool {
		return req.Transaction.SVMTransactionResponse.CuLimit.Equal(*originalCuLimit) &&
			req.Transaction.SVMTransactionResponse.CuPrice.Equal(multipliedCuPrice)
	})

	sender := registerSender(tc, ctx)
	expectSave(tc, ctx, func(saved *svmtransaction.SvmTransaction) bool {
		return saved.CuPrice.Equal(multipliedCuPrice)
	})
	expectPublish(tc, ctx, cmd.TransactionHash)

	resp, err := tc.handler.Handle(ctx, cmd)

	require.NoError(t, err)
	assert.Equal(t, validTxId, resp.NewTxId)
	sender.AssertExpectations(t)
}

func TestDltIngressRetrySvmTransactionCommandHandler_Handle_TxNotFound(t *testing.T) {
	tc := newTestContext()
	ctx := context.Background()
	cmd := &Command{TransactionHash: "missing-tx"}

	tc.transactionRepository.On("FindByTxId", ctx, "missing-tx").
		Return(nil, domainerrors.NewEntityNotFoundDomainError("Transaction", "missing-tx"))

	resp, err := tc.handler.Handle(ctx, cmd)

	assert.Nil(t, resp)
	assert.Error(t, err)
	tc.boundedBlockingQueue.AssertNotCalled(t, "Put", mock.Anything, mock.Anything, mock.Anything)
}

func TestDltIngressRetrySvmTransactionCommandHandler_Handle_CustodyKeyNotFound_ReleasesQueueSlot(t *testing.T) {
	cuLimit, _ := amount.NewFromString("200000")
	cuPrice, _ := amount.NewFromString("5000")
	oldTx := buildOldTx(t, cuLimit, cuPrice)

	tc := newTestContext()
	ctx := context.Background()
	cmd := &Command{TransactionHash: oldTx.TxId}

	tc.boundedBlockingQueue.On("Put", ctx, validNetworkId, mock.Anything).Return(nil)
	tc.boundedBlockingQueue.On("Take", ctx, validNetworkId).Return(nil)
	expectFindByTxId(tc, ctx, oldTx)
	tc.custodyKeyExistMultipleService.On("Execute", ctx, []string{oldTx.FeePayer}).
		Return(nil, domainerrors.NewEntityNotFoundDomainError("CustodyKey", oldTx.FeePayer))

	resp, err := tc.handler.Handle(ctx, cmd)

	assert.Nil(t, resp)
	assert.Error(t, err)
	tc.boundedBlockingQueue.AssertExpectations(t)
	tc.custodyProvider.AssertNotCalled(t, "Sign", mock.Anything, mock.Anything)
}

func TestDltIngressRetrySvmTransactionCommandHandler_Handle_SignFailure_ReleasesQueueSlot(t *testing.T) {
	cuLimit, _ := amount.NewFromString("200000")
	cuPrice, _ := amount.NewFromString("5000")
	oldTx := buildOldTx(t, cuLimit, cuPrice)

	custodyKeyFactory := custodykeyfactory.NewCustodyKeyTestFactory()
	custodyKey := custodyKeyFactory.CreateEntity(func(k *custodykey.CustodyKey) {
		k.DltAccountId = oldTx.FeePayer
	})

	tc := newTestContext()
	ctx := context.Background()
	cmd := &Command{TransactionHash: oldTx.TxId}

	tc.boundedBlockingQueue.On("Put", ctx, validNetworkId, mock.Anything).Return(nil)
	tc.boundedBlockingQueue.On("Take", ctx, validNetworkId).Return(nil)
	expectFindByTxId(tc, ctx, oldTx)
	expectSingleSignerCustodyKey(tc, ctx, oldTx.FeePayer, custodyKey)
	expectFreshBlockhash(tc, ctx)

	tc.custodyProvider.On("Sign", ctx, mock.Anything).Return(nil, assert.AnError)

	resp, err := tc.handler.Handle(ctx, cmd)

	assert.Nil(t, resp)
	assert.Error(t, err)
	tc.boundedBlockingQueue.AssertExpectations(t)
	tc.transactionRepository.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}

func TestDltIngressRetrySvmTransactionCommandHandler_Handle_SendFailure_HardDeletesPersistedTxAndReleasesQueueSlot(t *testing.T) {
	cuLimit, _ := amount.NewFromString("200000")
	cuPrice, _ := amount.NewFromString("5000")
	oldTx := buildOldTx(t, cuLimit, cuPrice)

	custodyKeyFactory := custodykeyfactory.NewCustodyKeyTestFactory()
	custodyKey := custodyKeyFactory.CreateEntity(func(k *custodykey.CustodyKey) {
		k.DltAccountId = oldTx.FeePayer
	})

	tc := newTestContext()
	ctx := context.Background()
	cmd := &Command{TransactionHash: oldTx.TxId}

	tc.boundedBlockingQueue.On("Put", ctx, validNetworkId, mock.Anything).Return(nil)
	tc.boundedBlockingQueue.On("Take", ctx, validNetworkId).Return(nil)
	expectFindByTxId(tc, ctx, oldTx)
	expectSingleSignerCustodyKey(tc, ctx, oldTx.FeePayer, custodyKey)
	expectFreshBlockhash(tc, ctx)
	expectSign(tc, ctx, func(req *custody.SignRequest) bool { return true })

	var savedTx *svmtransaction.SvmTransaction
	tc.transactionRepository.
		On("Save", context.Background(), mock.MatchedBy(func(saved *svmtransaction.SvmTransaction) bool {
			savedTx = saved
			return true
		})).
		Return(nil)
	tc.transactionRepository.
		On("HardDelete", context.Background(), mock.MatchedBy(func(deleted *svmtransaction.SvmTransaction) bool {
			return deleted == savedTx
		})).
		Return(nil)

	transactionSender := new(tsmock.TransactionSenderMock)
	tc.transactionSenderRegistry.Register(common.SVM, transactionSender)
	transactionSender.On("SendTransaction", ctx, mock.Anything).Return(nil, assert.AnError)

	resp, err := tc.handler.Handle(ctx, cmd)

	assert.Nil(t, resp)
	assert.Error(t, err)
	tc.boundedBlockingQueue.AssertExpectations(t)
	tc.transactionRepository.AssertExpectations(t)
	tc.eventBus.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
}
