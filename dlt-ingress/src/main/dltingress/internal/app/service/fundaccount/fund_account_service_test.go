package fundaccount

import (
	"context"
	"testing"

	mock2 "dlt-ingress/src/main/dltingress/internal/infra/evm/mock"
	"dlt-ingress/src/main/dltingress/internal/infra/repository/custodykey/mock"
	"dlt-ingress/src/main/dltingress/internal/infra/repository/faucetwallet/mock"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/app/command/signandsend"
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
	commandbusmock "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/test/core/command"
)

var (
	evmNetworkId       = "ethereum-sepolia"
	svmNetworkId       = "solana-devnet"
	hashgraphNetworkId = "hedera-testnet"
)

const targetDltAccountId = "0x1234567890AbCdEf1234567890AbCdEf1234ABcD"

func newAppService() (*AppService, *mockfaucetwalletrepo.Postgres, *mockcustodykeyrepo.Postgres, *mock2.EvmClientRegistryMock, *svmmocks.SvmClientRegistryMock, *commandbusmock.CommandBusMock) {
	config.DltIngressConfig = &config.Config{}
	config.DltIngressConfig.DltIngress.Networks = []*config.NetworkConfig{
		{Id: evmNetworkId, Dlt: string(common.EVM)},
	}

	faucetWalletRepo := new(mockfaucetwalletrepo.Postgres)
	custodyKeyRepo := new(mockcustodykeyrepo.Postgres)
	evmClientRegistry := new(mock2.EvmClientRegistryMock)
	svmClientRegistry := new(svmmocks.SvmClientRegistryMock)
	commandBus := new(commandbusmock.CommandBusMock)

	return NewAppService(faucetWalletRepo, custodyKeyRepo, evmClientRegistry, svmClientRegistry, commandBus),
		faucetWalletRepo, custodyKeyRepo, evmClientRegistry, svmClientRegistry, commandBus
}

func targetWalletKey(dlt common.Dlt) *custodykey.CustodyKey {
	return testcustodykey.NewCustodyKeyTestFactory().CreateEntity(func(k *custodykey.CustodyKey) {
		k.DltAccountId = targetDltAccountId
		k.Dlt = dlt
	})
}

func matchesFundCommand(networkId, senderDltAccountId, toDltAccountId string, amountToFund amount.Amount) any {
	return mock.MatchedBy(func(cmd *signandsend.Command) bool {
		return cmd.NetworkId == networkId &&
			cmd.SenderDltAccountId == senderDltAccountId &&
			len(cmd.SignersDltAccountIds) == 1 && cmd.SignersDltAccountIds[0] == senderDltAccountId &&
			cmd.MethodName == "nativeTransfer" &&
			cmd.MethodArgs["to"] == toDltAccountId &&
			cmd.MethodArgs["amount"] == amountToFund.String()
	})
}

func TestDltIngressFundAccountAppService_Execute_FindAccountToFundError(t *testing.T) {
	ctx := context.Background()
	service, _, custodyKeyRepo, _, _, _ := newAppService()

	custodyKeyRepo.On("FindByDltAccountId", ctx, targetDltAccountId).Return((*custodykey.CustodyKey)(nil), assert.AnError)

	err := service.Execute(ctx, &Request{DltAccountId: targetDltAccountId})

	assert.ErrorIs(t, err, assert.AnError)
}

func TestDltIngressFundAccountAppService_Execute_NoNetworkId_NoMatchingNetworks(t *testing.T) {
	ctx := context.Background()
	service, faucetWalletRepo, custodyKeyRepo, _, _, commandBus := newAppService()
	accountToFund := targetWalletKey(common.SVM)

	custodyKeyRepo.On("FindByDltAccountId", ctx, targetDltAccountId).Return(accountToFund, nil)

	err := service.Execute(ctx, &Request{DltAccountId: targetDltAccountId})

	assert.Nil(t, err)
	faucetWalletRepo.AssertNotCalled(t, "FindByNetworkId", mock.Anything, mock.Anything)
	commandBus.AssertNotCalled(t, "Dispatch", mock.Anything, mock.Anything)
}

func TestDltIngressFundAccountAppService_Execute_FaucetWalletNotFound(t *testing.T) {
	ctx := context.Background()
	service, faucetWalletRepo, custodyKeyRepo, _, _, commandBus := newAppService()
	accountToFund := targetWalletKey(common.EVM)

	custodyKeyRepo.On("FindByDltAccountId", ctx, targetDltAccountId).Return(accountToFund, nil)
	faucetWalletRepo.On("FindByNetworkId", ctx, evmNetworkId).
		Return((*faucetwallet.FaucetWallet)(nil), domainerrors.NewFaucetWalletNotFoundByNetworkIdDomainError(evmNetworkId))

	err := service.Execute(ctx, &Request{NetworkId: &evmNetworkId, DltAccountId: targetDltAccountId})

	assert.Nil(t, err)
	commandBus.AssertNotCalled(t, "Dispatch", mock.Anything, mock.Anything)
}

func TestDltIngressFundAccountAppService_Execute_NoNetworkId_FindFaucetWalletError(t *testing.T) {
	ctx := context.Background()
	service, faucetWalletRepo, custodyKeyRepo, _, _, _ := newAppService()
	accountToFund := targetWalletKey(common.EVM)

	custodyKeyRepo.On("FindByDltAccountId", ctx, targetDltAccountId).Return(accountToFund, nil)
	faucetWalletRepo.On("FindByNetworkId", ctx, evmNetworkId).Return((*faucetwallet.FaucetWallet)(nil), assert.AnError)

	err := service.Execute(ctx, &Request{DltAccountId: targetDltAccountId})

	assert.ErrorIs(t, err, assert.AnError)
}

func TestDltIngressFundAccountAppService_Execute_UnknownNetwork(t *testing.T) {
	ctx := context.Background()
	service, faucetWalletRepo, custodyKeyRepo, _, _, _ := newAppService()
	accountToFund := targetWalletKey(common.EVM)
	unknownNetworkId := "unknown-network"

	custodyKeyRepo.On("FindByDltAccountId", ctx, targetDltAccountId).Return(accountToFund, nil)

	err := service.Execute(ctx, &Request{NetworkId: &unknownNetworkId, DltAccountId: targetDltAccountId})

	assert.NotNil(t, err)
	assert.Equal(t, coreerror.ErrorCode("NETWORK_NOT_FOUND"), err.(coreerror.DomainError).ErrorCode())
	faucetWalletRepo.AssertNotCalled(t, "FindByNetworkId", mock.Anything, mock.Anything)
}

func TestDltIngressFundAccountAppService_Execute_DltMismatch(t *testing.T) {
	ctx := context.Background()
	service, faucetWalletRepo, custodyKeyRepo, _, _, _ := newAppService()
	accountToFund := targetWalletKey(common.SVM)

	custodyKeyRepo.On("FindByDltAccountId", ctx, targetDltAccountId).Return(accountToFund, nil)

	err := service.Execute(ctx, &Request{NetworkId: &evmNetworkId, DltAccountId: targetDltAccountId})

	assert.NotNil(t, err)
	assert.Equal(t, coreerror.ErrorCode("DLT_ACCOUNT_ID_NOT_VALID_FOR_NETWORK"), err.(coreerror.DomainError).ErrorCode())
	faucetWalletRepo.AssertNotCalled(t, "FindByNetworkId", mock.Anything, mock.Anything)
}

func TestDltIngressFundAccountAppService_Execute_FaucetWalletDisabled(t *testing.T) {
	ctx := context.Background()
	service, faucetWalletRepo, custodyKeyRepo, _, _, commandBus := newAppService()
	accountToFund := targetWalletKey(common.EVM)
	faucetWallet := testfaucetwallet.NewFaucetWalletTestFactory().CreateEntity(func(w *faucetwallet.FaucetWallet) {
		w.NetworkId = evmNetworkId
		w.Enabled = false
	})

	custodyKeyRepo.On("FindByDltAccountId", ctx, targetDltAccountId).Return(accountToFund, nil)
	faucetWalletRepo.On("FindByNetworkId", ctx, evmNetworkId).Return(faucetWallet, nil)

	err := service.Execute(ctx, &Request{NetworkId: &evmNetworkId, DltAccountId: targetDltAccountId})

	assert.Nil(t, err)
	custodyKeyRepo.AssertNotCalled(t, "FindById", mock.Anything, mock.Anything)
	commandBus.AssertNotCalled(t, "Dispatch", mock.Anything, mock.Anything)
}

func TestDltIngressFundAccountAppService_Execute_AccountBalanceAboveThreshold(t *testing.T) {
	ctx := context.Background()
	service, faucetWalletRepo, custodyKeyRepo, evmClientRegistry, _, commandBus := newAppService()
	accountToFund := targetWalletKey(common.EVM)
	faucetKey := testcustodykey.NewCustodyKeyTestFactory().CreateEntity()
	faucetWallet := testfaucetwallet.NewFaucetWalletTestFactory().CreateEntity(func(w *faucetwallet.FaucetWallet) {
		w.NetworkId = evmNetworkId
		w.CustodyKeyId = faucetKey.Id
	})
	evmClient := new(mock2.EvmClientMock)

	custodyKeyRepo.On("FindByDltAccountId", ctx, targetDltAccountId).Return(accountToFund, nil)
	faucetWalletRepo.On("FindByNetworkId", ctx, evmNetworkId).Return(faucetWallet, nil)
	custodyKeyRepo.On("FindById", ctx, faucetKey.Id).Return(faucetKey, nil)
	evmClientRegistry.On("GetClientForNetworkId", ctx, evmNetworkId).Return(evmClient, nil)
	accountToFundBalance := faucetWallet.BalanceThreshold.Add(*amount.One())
	evmClient.On("GetBalance", ctx, targetDltAccountId).Return(&accountToFundBalance, nil)

	err := service.Execute(ctx, &Request{NetworkId: &evmNetworkId, DltAccountId: targetDltAccountId})

	assert.Nil(t, err)
	faucetWalletRepo.AssertExpectations(t)
	custodyKeyRepo.AssertExpectations(t)
	evmClient.AssertExpectations(t)
	evmClient.AssertNotCalled(t, "GetBalance", ctx, faucetKey.DltAccountId)
	commandBus.AssertNotCalled(t, "Dispatch", mock.Anything, mock.Anything)
}

func TestDltIngressFundAccountAppService_Execute_FaucetWalletCustodyKeyError(t *testing.T) {
	ctx := context.Background()
	service, faucetWalletRepo, custodyKeyRepo, _, _, _ := newAppService()
	accountToFund := targetWalletKey(common.EVM)
	faucetKey := testcustodykey.NewCustodyKeyTestFactory().CreateEntity()
	faucetWallet := testfaucetwallet.NewFaucetWalletTestFactory().CreateEntity(func(w *faucetwallet.FaucetWallet) {
		w.NetworkId = evmNetworkId
		w.CustodyKeyId = faucetKey.Id
	})

	custodyKeyRepo.On("FindByDltAccountId", ctx, targetDltAccountId).Return(accountToFund, nil)
	faucetWalletRepo.On("FindByNetworkId", ctx, evmNetworkId).Return(faucetWallet, nil)
	custodyKeyRepo.On("FindById", ctx, faucetKey.Id).Return((*custodykey.CustodyKey)(nil), assert.AnError)

	err := service.Execute(ctx, &Request{NetworkId: &evmNetworkId, DltAccountId: targetDltAccountId})

	assert.ErrorIs(t, err, assert.AnError)
}

func TestDltIngressFundAccountAppService_Execute_DestinationBalanceError(t *testing.T) {
	ctx := context.Background()
	service, faucetWalletRepo, custodyKeyRepo, evmClientRegistry, _, _ := newAppService()
	accountToFund := targetWalletKey(common.EVM)
	faucetKey := testcustodykey.NewCustodyKeyTestFactory().CreateEntity()
	faucetWallet := testfaucetwallet.NewFaucetWalletTestFactory().CreateEntity(func(w *faucetwallet.FaucetWallet) {
		w.NetworkId = evmNetworkId
		w.CustodyKeyId = faucetKey.Id
	})
	evmClient := new(mock2.EvmClientMock)

	custodyKeyRepo.On("FindByDltAccountId", ctx, targetDltAccountId).Return(accountToFund, nil)
	faucetWalletRepo.On("FindByNetworkId", ctx, evmNetworkId).Return(faucetWallet, nil)
	custodyKeyRepo.On("FindById", ctx, faucetKey.Id).Return(faucetKey, nil)
	evmClientRegistry.On("GetClientForNetworkId", ctx, evmNetworkId).Return(evmClient, nil)
	evmClient.On("GetBalance", ctx, targetDltAccountId).Return(nil, assert.AnError)

	err := service.Execute(ctx, &Request{NetworkId: &evmNetworkId, DltAccountId: targetDltAccountId})

	assert.ErrorIs(t, err, assert.AnError)
	evmClient.AssertNotCalled(t, "GetBalance", ctx, faucetKey.DltAccountId)
}

func TestDltIngressFundAccountAppService_Execute_FaucetWalletInsufficientFunds(t *testing.T) {
	ctx := context.Background()
	service, faucetWalletRepo, custodyKeyRepo, evmClientRegistry, _, commandBus := newAppService()
	accountToFund := targetWalletKey(common.EVM)
	faucetKey := testcustodykey.NewCustodyKeyTestFactory().CreateEntity()
	faucetWallet := testfaucetwallet.NewFaucetWalletTestFactory().CreateEntity(func(w *faucetwallet.FaucetWallet) {
		w.NetworkId = evmNetworkId
		w.CustodyKeyId = faucetKey.Id
	})
	evmClient := new(mock2.EvmClientMock)

	custodyKeyRepo.On("FindByDltAccountId", ctx, targetDltAccountId).Return(accountToFund, nil)
	faucetWalletRepo.On("FindByNetworkId", ctx, evmNetworkId).Return(faucetWallet, nil)
	custodyKeyRepo.On("FindById", ctx, faucetKey.Id).Return(faucetKey, nil)
	evmClientRegistry.On("GetClientForNetworkId", ctx, evmNetworkId).Return(evmClient, nil)
	evmClient.On("GetBalance", ctx, targetDltAccountId).Return(amount.Zero(), nil)
	evmClient.On("GetBalance", ctx, faucetKey.DltAccountId).Return(amount.Zero(), nil)

	err := service.Execute(ctx, &Request{NetworkId: &evmNetworkId, DltAccountId: targetDltAccountId})

	assert.NotNil(t, err)
	assert.Equal(t, domainerrors.ErrorCodeFaucetWalletInsufficientFunds, err.(coreerror.DomainError).ErrorCode())
	evmClient.AssertExpectations(t)
	commandBus.AssertNotCalled(t, "Dispatch", mock.Anything, mock.Anything)
}

func TestDltIngressFundAccountAppService_Execute_FaucetWalletBalanceError(t *testing.T) {
	ctx := context.Background()
	service, faucetWalletRepo, custodyKeyRepo, evmClientRegistry, _, _ := newAppService()
	accountToFund := targetWalletKey(common.EVM)
	faucetKey := testcustodykey.NewCustodyKeyTestFactory().CreateEntity()
	faucetWallet := testfaucetwallet.NewFaucetWalletTestFactory().CreateEntity(func(w *faucetwallet.FaucetWallet) {
		w.NetworkId = evmNetworkId
		w.CustodyKeyId = faucetKey.Id
	})
	evmClient := new(mock2.EvmClientMock)

	custodyKeyRepo.On("FindByDltAccountId", ctx, targetDltAccountId).Return(accountToFund, nil)
	faucetWalletRepo.On("FindByNetworkId", ctx, evmNetworkId).Return(faucetWallet, nil)
	custodyKeyRepo.On("FindById", ctx, faucetKey.Id).Return(faucetKey, nil)
	evmClientRegistry.On("GetClientForNetworkId", ctx, evmNetworkId).Return(evmClient, nil)
	evmClient.On("GetBalance", ctx, targetDltAccountId).Return(amount.Zero(), nil)
	evmClient.On("GetBalance", ctx, faucetKey.DltAccountId).Return(nil, assert.AnError)

	err := service.Execute(ctx, &Request{NetworkId: &evmNetworkId, DltAccountId: targetDltAccountId})

	assert.ErrorIs(t, err, assert.AnError)
}

func TestDltIngressFundAccountAppService_Execute_EvmClientRegistryError(t *testing.T) {
	ctx := context.Background()
	service, faucetWalletRepo, custodyKeyRepo, evmClientRegistry, _, _ := newAppService()
	accountToFund := targetWalletKey(common.EVM)
	faucetKey := testcustodykey.NewCustodyKeyTestFactory().CreateEntity()
	faucetWallet := testfaucetwallet.NewFaucetWalletTestFactory().CreateEntity(func(w *faucetwallet.FaucetWallet) {
		w.NetworkId = evmNetworkId
		w.CustodyKeyId = faucetKey.Id
	})

	custodyKeyRepo.On("FindByDltAccountId", ctx, targetDltAccountId).Return(accountToFund, nil)
	faucetWalletRepo.On("FindByNetworkId", ctx, evmNetworkId).Return(faucetWallet, nil)
	custodyKeyRepo.On("FindById", ctx, faucetKey.Id).Return(faucetKey, nil)
	evmClientRegistry.On("GetClientForNetworkId", ctx, evmNetworkId).Return(nil, assert.AnError)

	err := service.Execute(ctx, &Request{NetworkId: &evmNetworkId, DltAccountId: targetDltAccountId})

	assert.ErrorIs(t, err, assert.AnError)
}

func TestDltIngressFundAccountAppService_Execute_SvmClientRegistryError(t *testing.T) {
	ctx := context.Background()
	service, faucetWalletRepo, custodyKeyRepo, _, svmClientRegistry, _ := newAppService()
	config.DltIngressConfig.DltIngress.Networks = append(config.DltIngressConfig.DltIngress.Networks, &config.NetworkConfig{Id: svmNetworkId, Dlt: string(common.SVM)})

	accountToFund := targetWalletKey(common.SVM)
	faucetKey := testcustodykey.NewCustodyKeyTestFactory().CreateEntity(func(k *custodykey.CustodyKey) { k.Dlt = common.SVM })
	faucetWallet := testfaucetwallet.NewFaucetWalletTestFactory().CreateEntity(func(w *faucetwallet.FaucetWallet) {
		w.NetworkId = svmNetworkId
		w.CustodyKeyId = faucetKey.Id
	})

	custodyKeyRepo.On("FindByDltAccountId", ctx, targetDltAccountId).Return(accountToFund, nil)
	faucetWalletRepo.On("FindByNetworkId", ctx, svmNetworkId).Return(faucetWallet, nil)
	custodyKeyRepo.On("FindById", ctx, faucetKey.Id).Return(faucetKey, nil)
	svmClientRegistry.On("GetClientForNetworkId", svmNetworkId).Return(nil, assert.AnError)

	err := service.Execute(ctx, &Request{NetworkId: &svmNetworkId, DltAccountId: targetDltAccountId})

	assert.ErrorIs(t, err, assert.AnError)
}

func TestDltIngressFundAccountAppService_Execute_UnsupportedDlt(t *testing.T) {
	ctx := context.Background()
	service, faucetWalletRepo, custodyKeyRepo, _, _, _ := newAppService()
	config.DltIngressConfig.DltIngress.Networks = append(config.DltIngressConfig.DltIngress.Networks, &config.NetworkConfig{Id: hashgraphNetworkId, Dlt: string(common.Hashgraph)})

	accountToFund := testcustodykey.NewCustodyKeyTestFactory().CreateEntity(func(k *custodykey.CustodyKey) {
		k.DltAccountId = targetDltAccountId
		k.Dlt = common.Hashgraph
	})
	faucetKey := testcustodykey.NewCustodyKeyTestFactory().CreateEntity(func(k *custodykey.CustodyKey) { k.Dlt = common.Hashgraph })
	faucetWallet := testfaucetwallet.NewFaucetWalletTestFactory().CreateEntity(func(w *faucetwallet.FaucetWallet) {
		w.NetworkId = hashgraphNetworkId
		w.CustodyKeyId = faucetKey.Id
	})

	custodyKeyRepo.On("FindByDltAccountId", ctx, targetDltAccountId).Return(accountToFund, nil)
	faucetWalletRepo.On("FindByNetworkId", ctx, hashgraphNetworkId).Return(faucetWallet, nil)
	custodyKeyRepo.On("FindById", ctx, faucetKey.Id).Return(faucetKey, nil)

	err := service.Execute(ctx, &Request{NetworkId: &hashgraphNetworkId, DltAccountId: targetDltAccountId})

	assert.NotNil(t, err)
	assert.Equal(t, domainerrors.ErrorCodeInvalidDlt, err.(coreerror.DomainError).ErrorCode())
}

func TestDltIngressFundAccountAppService_Execute_DispatchError(t *testing.T) {
	ctx := context.Background()
	service, faucetWalletRepo, custodyKeyRepo, evmClientRegistry, _, commandBus := newAppService()
	accountToFund := targetWalletKey(common.EVM)
	faucetKey := testcustodykey.NewCustodyKeyTestFactory().CreateEntity()
	faucetWallet := testfaucetwallet.NewFaucetWalletTestFactory().CreateEntity(func(w *faucetwallet.FaucetWallet) {
		w.NetworkId = evmNetworkId
		w.CustodyKeyId = faucetKey.Id
	})
	evmClient := new(mock2.EvmClientMock)
	destinationBalance := *amount.Zero()
	amountToFund := faucetWallet.FundingAmount.Sub(destinationBalance)

	custodyKeyRepo.On("FindByDltAccountId", ctx, targetDltAccountId).Return(accountToFund, nil)
	faucetWalletRepo.On("FindByNetworkId", ctx, evmNetworkId).Return(faucetWallet, nil)
	custodyKeyRepo.On("FindById", ctx, faucetKey.Id).Return(faucetKey, nil)
	evmClientRegistry.On("GetClientForNetworkId", ctx, evmNetworkId).Return(evmClient, nil)
	evmClient.On("GetBalance", ctx, faucetKey.DltAccountId).Return(amount.Ten(), nil)
	evmClient.On("GetBalance", ctx, targetDltAccountId).Return(&destinationBalance, nil)
	commandBus.On("Dispatch", ctx, matchesFundCommand(evmNetworkId, faucetKey.DltAccountId, targetDltAccountId, amountToFund)).Return(nil, assert.AnError)

	err := service.Execute(ctx, &Request{NetworkId: &evmNetworkId, DltAccountId: targetDltAccountId})

	assert.ErrorIs(t, err, assert.AnError)
}

func TestDltIngressFundAccountAppService_Execute_EvmSuccess(t *testing.T) {
	ctx := context.Background()
	service, faucetWalletRepo, custodyKeyRepo, evmClientRegistry, _, commandBus := newAppService()
	accountToFund := targetWalletKey(common.EVM)
	faucetKey := testcustodykey.NewCustodyKeyTestFactory().CreateEntity()
	faucetWallet := testfaucetwallet.NewFaucetWalletTestFactory().CreateEntity(func(w *faucetwallet.FaucetWallet) {
		w.NetworkId = evmNetworkId
		w.CustodyKeyId = faucetKey.Id
	})
	evmClient := new(mock2.EvmClientMock)
	destinationBalance := *amount.Zero()
	amountToFund := faucetWallet.FundingAmount.Sub(destinationBalance)

	custodyKeyRepo.On("FindByDltAccountId", ctx, targetDltAccountId).Return(accountToFund, nil)
	faucetWalletRepo.On("FindByNetworkId", ctx, evmNetworkId).Return(faucetWallet, nil)
	custodyKeyRepo.On("FindById", ctx, faucetKey.Id).Return(faucetKey, nil)
	evmClientRegistry.On("GetClientForNetworkId", ctx, evmNetworkId).Return(evmClient, nil)
	evmClient.On("GetBalance", ctx, faucetKey.DltAccountId).Return(amount.Ten(), nil)
	evmClient.On("GetBalance", ctx, targetDltAccountId).Return(&destinationBalance, nil)
	commandBus.On("Dispatch", ctx, matchesFundCommand(evmNetworkId, faucetKey.DltAccountId, targetDltAccountId, amountToFund)).Return(nil, nil)

	err := service.Execute(ctx, &Request{NetworkId: &evmNetworkId, DltAccountId: targetDltAccountId})

	assert.Nil(t, err)
	faucetWalletRepo.AssertExpectations(t)
	custodyKeyRepo.AssertExpectations(t)
	evmClient.AssertExpectations(t)
	commandBus.AssertExpectations(t)
}

func TestDltIngressFundAccountAppService_Execute_SvmSuccess(t *testing.T) {
	ctx := context.Background()
	service, faucetWalletRepo, custodyKeyRepo, _, svmClientRegistry, commandBus := newAppService()
	config.DltIngressConfig.DltIngress.Networks = append(config.DltIngressConfig.DltIngress.Networks, &config.NetworkConfig{Id: svmNetworkId, Dlt: string(common.SVM)})
	accountToFund := targetWalletKey(common.SVM)
	faucetKey := testcustodykey.NewCustodyKeyTestFactory().CreateEntity(func(k *custodykey.CustodyKey) { k.Dlt = common.SVM })
	faucetWallet := testfaucetwallet.NewFaucetWalletTestFactory().CreateEntity(func(w *faucetwallet.FaucetWallet) {
		w.NetworkId = svmNetworkId
		w.CustodyKeyId = faucetKey.Id
	})
	svmClient := new(svmmocks.SvmClientMock)
	destinationBalance := *amount.Zero()
	amountToFund := faucetWallet.FundingAmount.Sub(destinationBalance)

	custodyKeyRepo.On("FindByDltAccountId", ctx, targetDltAccountId).Return(accountToFund, nil)
	faucetWalletRepo.On("FindByNetworkId", ctx, svmNetworkId).Return(faucetWallet, nil)
	custodyKeyRepo.On("FindById", ctx, faucetKey.Id).Return(faucetKey, nil)
	svmClientRegistry.On("GetClientForNetworkId", svmNetworkId).Return(svmClient, nil)
	svmClient.On("GetBalance", ctx, faucetKey.DltAccountId).Return(amount.Ten(), nil)
	svmClient.On("GetBalance", ctx, targetDltAccountId).Return(&destinationBalance, nil)
	commandBus.On("Dispatch", ctx, matchesFundCommand(svmNetworkId, faucetKey.DltAccountId, targetDltAccountId, amountToFund)).Return(nil, nil)

	err := service.Execute(ctx, &Request{NetworkId: &svmNetworkId, DltAccountId: targetDltAccountId})

	assert.Nil(t, err)
	faucetWalletRepo.AssertExpectations(t)
	custodyKeyRepo.AssertExpectations(t)
	svmClient.AssertExpectations(t)
	commandBus.AssertExpectations(t)
}

func TestDltIngressFundAccountAppService_Execute_PartialTopUp(t *testing.T) {
	ctx := context.Background()
	service, faucetWalletRepo, custodyKeyRepo, evmClientRegistry, _, commandBus := newAppService()
	accountToFund := targetWalletKey(common.EVM)
	faucetKey := testcustodykey.NewCustodyKeyTestFactory().CreateEntity()
	faucetWallet := testfaucetwallet.NewFaucetWalletTestFactory().CreateEntity(func(w *faucetwallet.FaucetWallet) {
		fundingAmount, _ := amount.NewFromString("10")
		balanceThreshold, _ := amount.NewFromString("8")

		w.NetworkId = evmNetworkId
		w.CustodyKeyId = faucetKey.Id
		w.FundingAmount = *fundingAmount
		w.BalanceThreshold = *balanceThreshold
	})
	evmClient := new(mock2.EvmClientMock)

	// destinationBalance (3) < BalanceThreshold (8) => amountToFund = FundingAmount (10) - destinationBalance (3) = 7.
	// faucetWalletBalance (8) < amountToFund (7)
	destinationBalance, _ := amount.NewFromString("3")
	faucetBalance, _ := amount.NewFromString("8")
	amountToFund := faucetWallet.FundingAmount.Sub(*destinationBalance)

	custodyKeyRepo.On("FindByDltAccountId", ctx, targetDltAccountId).Return(accountToFund, nil)
	faucetWalletRepo.On("FindByNetworkId", ctx, evmNetworkId).Return(faucetWallet, nil)
	custodyKeyRepo.On("FindById", ctx, faucetKey.Id).Return(faucetKey, nil)
	evmClientRegistry.On("GetClientForNetworkId", ctx, evmNetworkId).Return(evmClient, nil)
	evmClient.On("GetBalance", ctx, targetDltAccountId).Return(destinationBalance, nil)
	evmClient.On("GetBalance", ctx, faucetKey.DltAccountId).Return(faucetBalance, nil)
	commandBus.On("Dispatch", ctx, matchesFundCommand(evmNetworkId, faucetKey.DltAccountId, targetDltAccountId, amountToFund)).Return(nil, nil)

	err := service.Execute(ctx, &Request{NetworkId: &evmNetworkId, DltAccountId: targetDltAccountId})

	assert.Nil(t, err)
	commandBus.AssertExpectations(t)
}

func TestDltIngressFundAccountAppService_Execute_NoNetworkId_FundsAllMatchingNetworks(t *testing.T) {
	ctx := context.Background()
	service, faucetWalletRepo, custodyKeyRepo, evmClientRegistry, _, commandBus := newAppService()
	secondEvmNetworkId := "polygon-amoy"
	config.DltIngressConfig.DltIngress.Networks = append(config.DltIngressConfig.DltIngress.Networks, &config.NetworkConfig{Id: secondEvmNetworkId, Dlt: string(common.EVM)})
	accountToFund := targetWalletKey(common.EVM)
	evmClient := new(mock2.EvmClientMock)

	faucetKey1 := testcustodykey.NewCustodyKeyTestFactory().CreateEntity()
	faucetWallet1 := testfaucetwallet.NewFaucetWalletTestFactory().CreateEntity(func(w *faucetwallet.FaucetWallet) {
		w.NetworkId = evmNetworkId
		w.CustodyKeyId = faucetKey1.Id
	})
	faucetKey2 := testcustodykey.NewCustodyKeyTestFactory().CreateEntity()
	faucetWallet2 := testfaucetwallet.NewFaucetWalletTestFactory().CreateEntity(func(w *faucetwallet.FaucetWallet) {
		w.NetworkId = secondEvmNetworkId
		w.CustodyKeyId = faucetKey2.Id
	})

	destinationBalance := *amount.Zero()
	amountToFund1 := faucetWallet1.FundingAmount.Sub(destinationBalance)
	amountToFund2 := faucetWallet2.FundingAmount.Sub(destinationBalance)

	custodyKeyRepo.On("FindByDltAccountId", ctx, targetDltAccountId).Return(accountToFund, nil)

	faucetWalletRepo.On("FindByNetworkId", ctx, evmNetworkId).Return(faucetWallet1, nil)
	custodyKeyRepo.On("FindById", ctx, faucetKey1.Id).Return(faucetKey1, nil)
	evmClientRegistry.On("GetClientForNetworkId", ctx, evmNetworkId).Return(evmClient, nil)
	evmClient.On("GetBalance", ctx, faucetKey1.DltAccountId).Return(amount.Ten(), nil)
	commandBus.On("Dispatch", ctx, matchesFundCommand(evmNetworkId, faucetKey1.DltAccountId, targetDltAccountId, amountToFund1)).Return(nil, nil)

	faucetWalletRepo.On("FindByNetworkId", ctx, secondEvmNetworkId).Return(faucetWallet2, nil)
	custodyKeyRepo.On("FindById", ctx, faucetKey2.Id).Return(faucetKey2, nil)
	evmClientRegistry.On("GetClientForNetworkId", ctx, secondEvmNetworkId).Return(evmClient, nil)
	evmClient.On("GetBalance", ctx, faucetKey2.DltAccountId).Return(amount.Ten(), nil)
	commandBus.On("Dispatch", ctx, matchesFundCommand(secondEvmNetworkId, faucetKey2.DltAccountId, targetDltAccountId, amountToFund2)).Return(nil, nil)

	evmClient.On("GetBalance", ctx, targetDltAccountId).Return(&destinationBalance, nil)

	err := service.Execute(ctx, &Request{DltAccountId: targetDltAccountId})

	assert.Nil(t, err)
	faucetWalletRepo.AssertExpectations(t)
	custodyKeyRepo.AssertExpectations(t)
	evmClient.AssertNumberOfCalls(t, "GetBalance", 4)
	commandBus.AssertNumberOfCalls(t, "Dispatch", 2)
}
