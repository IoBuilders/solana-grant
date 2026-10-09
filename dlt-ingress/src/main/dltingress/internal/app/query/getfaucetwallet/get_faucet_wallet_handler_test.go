package getfaucetwallet

import (
	"context"
	"testing"

	mock2 "dlt-ingress/src/main/dltingress/internal/infra/evm/mock"
	"dlt-ingress/src/main/dltingress/internal/infra/repository/custodykey/mock"
	"dlt-ingress/src/main/dltingress/internal/infra/repository/faucetwallet/mock"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"dlt-ingress/src/main/dltingress/internal/domain/custodykey"
	testcustodykey "dlt-ingress/src/main/dltingress/internal/domain/custodykey"
	"dlt-ingress/src/main/dltingress/internal/domain/domainerrors"
	"dlt-ingress/src/main/dltingress/internal/domain/faucetwallet"
	testfaucetwallet "dlt-ingress/src/main/dltingress/internal/domain/faucetwallet"
	svmmocks "dlt-ingress/src/main/dltingress/internal/infra/svm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
)

const (
	evmNetworkId = "ethereum-sepolia"
	svmNetworkId = "solana-devnet"
)

func newHandler() (*QueryHandler, *mockfaucetwalletrepo.Postgres, *mockcustodykeyrepo.Postgres, *mock2.EvmClientRegistryMock, *svmmocks.SvmClientRegistryMock) {
	config.DltIngressConfig = &config.Config{}
	config.DltIngressConfig.DltIngress.Networks = []*config.NetworkConfig{
		{Id: evmNetworkId, Dlt: string(common.EVM)},
		{Id: svmNetworkId, Dlt: string(common.SVM)},
	}

	faucetWalletRepo := new(mockfaucetwalletrepo.Postgres)
	custodyKeyRepo := new(mockcustodykeyrepo.Postgres)
	evmClientRegistry := new(mock2.EvmClientRegistryMock)
	svmClientRegistry := new(svmmocks.SvmClientRegistryMock)

	return NewHandler(faucetWalletRepo, custodyKeyRepo, evmClientRegistry, svmClientRegistry),
		faucetWalletRepo, custodyKeyRepo, evmClientRegistry, svmClientRegistry
}

func TestDltIngressGetFaucetWalletQueryHandler_Execute_EvmSuccess(t *testing.T) {
	handler, faucetWalletRepo, custodyKeyRepo, evmClientRegistry, _ := newHandler()
	walletFactory := testfaucetwallet.NewFaucetWalletTestFactory()
	keyFactory := testcustodykey.NewCustodyKeyTestFactory()

	key := keyFactory.CreateEntity()
	wallet := walletFactory.CreateEntity(func(w *faucetwallet.FaucetWallet) {
		w.NetworkId = evmNetworkId
		w.CustodyKeyId = key.Id
	})
	evmClient := new(mock2.EvmClientMock)

	faucetWalletRepo.On("FindByNetworkId", mock.Anything, evmNetworkId).Return(wallet, nil)
	custodyKeyRepo.On("FindById", mock.Anything, key.Id).Return(key, nil)
	evmClientRegistry.On("GetClientForNetworkId", mock.Anything, evmNetworkId).Return(evmClient, nil)
	evmClient.On("GetBalance", mock.Anything, key.DltAccountId).Return(amount.Ten(), nil)

	rawResp, err := handler.Execute(context.Background(), Query{NetworkId: evmNetworkId})

	assert.Nil(t, err)
	resp := rawResp.(Response)
	assert.Equal(t, wallet.Id, resp.FaucetWalletId)
	assert.Equal(t, evmNetworkId, resp.NetworkId)
	assert.Equal(t, key.DltAccountId, resp.DltAccountId)
	assert.True(t, wallet.FundingAmount.Equal(resp.FundingAmount))
	assert.True(t, wallet.BalanceThreshold.Equal(resp.BalanceThreshold))
	assert.True(t, amount.Ten().Equal(resp.Balance))
	assert.Equal(t, wallet.Enabled, resp.Enabled)
	faucetWalletRepo.AssertExpectations(t)
	custodyKeyRepo.AssertExpectations(t)
	evmClientRegistry.AssertExpectations(t)
	evmClient.AssertExpectations(t)
}

func TestDltIngressGetFaucetWalletQueryHandler_Execute_SvmSuccess(t *testing.T) {
	handler, faucetWalletRepo, custodyKeyRepo, _, svmClientRegistry := newHandler()
	walletFactory := testfaucetwallet.NewFaucetWalletTestFactory()
	keyFactory := testcustodykey.NewCustodyKeyTestFactory()

	key := keyFactory.CreateEntity(func(k *custodykey.CustodyKey) {
		k.Dlt = common.SVM
		k.DltAccountId = "5eykt4UsFv8P8NJdTREpY1vzqKqZKvdpKuc147dw2N9d"
	})
	wallet := walletFactory.CreateEntity(func(w *faucetwallet.FaucetWallet) {
		w.NetworkId = svmNetworkId
		w.CustodyKeyId = key.Id
	})
	svmClient := new(svmmocks.SvmClientMock)

	faucetWalletRepo.On("FindByNetworkId", mock.Anything, svmNetworkId).Return(wallet, nil)
	custodyKeyRepo.On("FindById", mock.Anything, key.Id).Return(key, nil)
	svmClientRegistry.On("GetClientForNetworkId", svmNetworkId).Return(svmClient, nil)
	svmClient.On("GetBalance", mock.Anything, key.DltAccountId).Return(amount.One(), nil)

	rawResp, err := handler.Execute(context.Background(), Query{NetworkId: svmNetworkId})

	assert.Nil(t, err)
	resp := rawResp.(Response)
	assert.Equal(t, key.DltAccountId, resp.DltAccountId)
	assert.True(t, amount.One().Equal(resp.Balance))
	faucetWalletRepo.AssertExpectations(t)
	custodyKeyRepo.AssertExpectations(t)
	svmClientRegistry.AssertExpectations(t)
	svmClient.AssertExpectations(t)
}

func TestDltIngressGetFaucetWalletQueryHandler_Execute_FaucetWalletNotFound(t *testing.T) {
	handler, faucetWalletRepo, custodyKeyRepo, _, _ := newHandler()
	expectedError := domainerrors.NewFaucetWalletNotFoundByNetworkIdDomainError(evmNetworkId)

	faucetWalletRepo.On("FindByNetworkId", mock.Anything, evmNetworkId).Return((*faucetwallet.FaucetWallet)(nil), expectedError)

	rawResp, err := handler.Execute(context.Background(), Query{NetworkId: evmNetworkId})

	assert.Nil(t, rawResp)
	assert.NotNil(t, err)
	assert.Equal(t, expectedError.Error(), err.Error())
	custodyKeyRepo.AssertNotCalled(t, "FindById", mock.Anything, mock.Anything)
}

func TestDltIngressGetFaucetWalletQueryHandler_Execute_CustodyKeyRepositoryError(t *testing.T) {
	handler, faucetWalletRepo, custodyKeyRepo, evmClientRegistry, _ := newHandler()
	walletFactory := testfaucetwallet.NewFaucetWalletTestFactory()
	wallet := walletFactory.CreateEntity(func(w *faucetwallet.FaucetWallet) { w.NetworkId = evmNetworkId })

	faucetWalletRepo.On("FindByNetworkId", mock.Anything, evmNetworkId).Return(wallet, nil)
	custodyKeyRepo.On("FindById", mock.Anything, wallet.CustodyKeyId).Return((*custodykey.CustodyKey)(nil), assert.AnError)

	rawResp, err := handler.Execute(context.Background(), Query{NetworkId: evmNetworkId})

	assert.Nil(t, rawResp)
	assert.ErrorIs(t, err, assert.AnError)
	evmClientRegistry.AssertNotCalled(t, "GetClientForNetworkId", mock.Anything, mock.Anything)
}

func TestDltIngressGetFaucetWalletQueryHandler_Execute_UnknownNetwork(t *testing.T) {
	handler, faucetWalletRepo, custodyKeyRepo, _, _ := newHandler()

	rawResp, err := handler.Execute(context.Background(), Query{NetworkId: "unknown-network"})

	assert.Nil(t, rawResp)
	assert.NotNil(t, err)
	faucetWalletRepo.AssertNotCalled(t, "FindByNetworkId", mock.Anything, mock.Anything)
	custodyKeyRepo.AssertNotCalled(t, "FindById", mock.Anything, mock.Anything)
}

func TestDltIngressGetFaucetWalletQueryHandler_Execute_InvalidDlt(t *testing.T) {
	handler, faucetWalletRepo, custodyKeyRepo, _, _ := newHandler()
	config.DltIngressConfig.DltIngress.Networks = append(config.DltIngressConfig.DltIngress.Networks, &config.NetworkConfig{Id: "hashgraph-testnet", Dlt: "INVALID"})

	rawResp, err := handler.Execute(context.Background(), Query{NetworkId: "hashgraph-testnet"})

	assert.Nil(t, rawResp)
	assert.NotNil(t, err)
	assert.Equal(t, domainerrors.ErrorCodeInvalidDlt, err.(coreerror.DomainError).ErrorCode())
	faucetWalletRepo.AssertNotCalled(t, "FindByNetworkId", mock.Anything, mock.Anything)
	custodyKeyRepo.AssertNotCalled(t, "FindById", mock.Anything, mock.Anything)
}

func TestDltIngressGetFaucetWalletQueryHandler_Execute_EvmClientRegistryError(t *testing.T) {
	handler, faucetWalletRepo, custodyKeyRepo, evmClientRegistry, _ := newHandler()
	walletFactory := testfaucetwallet.NewFaucetWalletTestFactory()
	keyFactory := testcustodykey.NewCustodyKeyTestFactory()

	key := keyFactory.CreateEntity()
	wallet := walletFactory.CreateEntity(func(w *faucetwallet.FaucetWallet) {
		w.NetworkId = evmNetworkId
		w.CustodyKeyId = key.Id
	})

	faucetWalletRepo.On("FindByNetworkId", mock.Anything, evmNetworkId).Return(wallet, nil)
	custodyKeyRepo.On("FindById", mock.Anything, key.Id).Return(key, nil)
	evmClientRegistry.On("GetClientForNetworkId", mock.Anything, evmNetworkId).Return(nil, assert.AnError)

	rawResp, err := handler.Execute(context.Background(), Query{NetworkId: evmNetworkId})

	assert.Nil(t, rawResp)
	assert.ErrorIs(t, err, assert.AnError)
}

func TestDltIngressGetFaucetWalletQueryHandler_Execute_EvmGetBalanceError(t *testing.T) {
	handler, faucetWalletRepo, custodyKeyRepo, evmClientRegistry, _ := newHandler()
	walletFactory := testfaucetwallet.NewFaucetWalletTestFactory()
	keyFactory := testcustodykey.NewCustodyKeyTestFactory()

	key := keyFactory.CreateEntity()
	wallet := walletFactory.CreateEntity(func(w *faucetwallet.FaucetWallet) {
		w.NetworkId = evmNetworkId
		w.CustodyKeyId = key.Id
	})
	evmClient := new(mock2.EvmClientMock)

	faucetWalletRepo.On("FindByNetworkId", mock.Anything, evmNetworkId).Return(wallet, nil)
	custodyKeyRepo.On("FindById", mock.Anything, key.Id).Return(key, nil)
	evmClientRegistry.On("GetClientForNetworkId", mock.Anything, evmNetworkId).Return(evmClient, nil)
	evmClient.On("GetBalance", mock.Anything, key.DltAccountId).Return(nil, assert.AnError)

	rawResp, err := handler.Execute(context.Background(), Query{NetworkId: evmNetworkId})

	assert.Nil(t, rawResp)
	assert.ErrorIs(t, err, assert.AnError)
}

func TestDltIngressGetFaucetWalletQueryHandler_Execute_SvmClientRegistryError(t *testing.T) {
	handler, faucetWalletRepo, custodyKeyRepo, _, svmClientRegistry := newHandler()
	walletFactory := testfaucetwallet.NewFaucetWalletTestFactory()
	keyFactory := testcustodykey.NewCustodyKeyTestFactory()

	key := keyFactory.CreateEntity(func(k *custodykey.CustodyKey) { k.Dlt = common.SVM })
	wallet := walletFactory.CreateEntity(func(w *faucetwallet.FaucetWallet) {
		w.NetworkId = svmNetworkId
		w.CustodyKeyId = key.Id
	})

	faucetWalletRepo.On("FindByNetworkId", mock.Anything, svmNetworkId).Return(wallet, nil)
	custodyKeyRepo.On("FindById", mock.Anything, key.Id).Return(key, nil)
	svmClientRegistry.On("GetClientForNetworkId", svmNetworkId).Return(nil, assert.AnError)

	rawResp, err := handler.Execute(context.Background(), Query{NetworkId: svmNetworkId})

	assert.Nil(t, rawResp)
	assert.ErrorIs(t, err, assert.AnError)
}

func TestDltIngressGetFaucetWalletQueryHandler_Execute_SvmGetBalanceError(t *testing.T) {
	handler, faucetWalletRepo, custodyKeyRepo, _, svmClientRegistry := newHandler()
	walletFactory := testfaucetwallet.NewFaucetWalletTestFactory()
	keyFactory := testcustodykey.NewCustodyKeyTestFactory()

	key := keyFactory.CreateEntity(func(k *custodykey.CustodyKey) { k.Dlt = common.SVM })
	wallet := walletFactory.CreateEntity(func(w *faucetwallet.FaucetWallet) {
		w.NetworkId = svmNetworkId
		w.CustodyKeyId = key.Id
	})
	svmClient := new(svmmocks.SvmClientMock)

	faucetWalletRepo.On("FindByNetworkId", mock.Anything, svmNetworkId).Return(wallet, nil)
	custodyKeyRepo.On("FindById", mock.Anything, key.Id).Return(key, nil)
	svmClientRegistry.On("GetClientForNetworkId", svmNetworkId).Return(svmClient, nil)
	svmClient.On("GetBalance", mock.Anything, key.DltAccountId).Return(nil, assert.AnError)

	rawResp, err := handler.Execute(context.Background(), Query{NetworkId: svmNetworkId})

	assert.Nil(t, rawResp)
	assert.ErrorIs(t, err, assert.AnError)
}

func TestDltIngressGetFaucetWalletQueryHandler_Execute_UnsupportedDlt(t *testing.T) {
	handler, faucetWalletRepo, custodyKeyRepo, _, _ := newHandler()
	const hashgraphNetworkId = "hedera-testnet"
	config.DltIngressConfig.DltIngress.Networks = append(config.DltIngressConfig.DltIngress.Networks, &config.NetworkConfig{Id: hashgraphNetworkId, Dlt: string(common.Hashgraph)})
	walletFactory := testfaucetwallet.NewFaucetWalletTestFactory()
	keyFactory := testcustodykey.NewCustodyKeyTestFactory()

	key := keyFactory.CreateEntity(func(k *custodykey.CustodyKey) { k.Dlt = common.Hashgraph })
	wallet := walletFactory.CreateEntity(func(w *faucetwallet.FaucetWallet) {
		w.NetworkId = hashgraphNetworkId
		w.CustodyKeyId = key.Id
	})

	faucetWalletRepo.On("FindByNetworkId", mock.Anything, hashgraphNetworkId).Return(wallet, nil)
	custodyKeyRepo.On("FindById", mock.Anything, key.Id).Return(key, nil)

	rawResp, err := handler.Execute(context.Background(), Query{NetworkId: hashgraphNetworkId})

	assert.Nil(t, rawResp)
	assert.NotNil(t, err)
	assert.Equal(t, domainerrors.ErrorCodeInvalidDlt, err.(coreerror.DomainError).ErrorCode())
}
