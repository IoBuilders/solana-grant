package createfaucetwallet

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/infra/repository/custodykey/mock"
	"dlt-ingress/src/main/dltingress/internal/infra/repository/faucetwallet/mock"
	"testing"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/domain/custodykey"
	"dlt-ingress/src/main/dltingress/internal/domain/domainerrors"
	"dlt-ingress/src/main/dltingress/internal/domain/faucetwallet"
	"dlt-ingress/src/main/dltingress/internal/infra/custody"
	custodymocks "dlt-ingress/src/main/dltingress/internal/infra/custody/mock"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
	testbus "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/test/core/event"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	validNetworkId    = "ethereum-sepolia"
	validDlt          = "EVM"
	validDltAccountId = "0xAbCdEf1234567890AbCdEf1234567890AbCdEf12"
	validExternalId   = "ext-id-123"
)

func newHandler() (*CommandHandler, *mockfaucetwalletrepo.Postgres, *mockcustodykeyrepo.Postgres, *custodymocks.CustodyProviderMock, *testbus.EventBusMock) {
	config.DltIngressConfig = &config.Config{}
	config.DltIngressConfig.DltIngress.Networks = []*config.NetworkConfig{
		{Id: validNetworkId, Dlt: validDlt},
	}
	config.DltIngressConfig.DltIngress.Custody.Provider = string(custodykey.CustodyProviderDFNS)

	faucetWalletRepo := new(mockfaucetwalletrepo.Postgres)
	custodyKeyRepo := new(mockcustodykeyrepo.Postgres)
	provider := new(custodymocks.CustodyProviderMock)
	eventBus := new(testbus.EventBusMock)

	return NewCommandHandler(faucetWalletRepo, custodyKeyRepo, provider, eventBus), faucetWalletRepo, custodyKeyRepo, provider, eventBus
}

func matchesCustodyKey(dltAccountId string) any {
	return mock.MatchedBy(func(k *custodykey.CustodyKey) bool {
		return k.DltAccountId == dltAccountId
	})
}

func matchesCreatedWallet(fundingAmount, balanceThreshold amount.Amount) any {
	return mock.MatchedBy(func(w *faucetwallet.FaucetWallet) bool {
		return w.NetworkId == validNetworkId &&
			fundingAmount.Equal(w.FundingAmount) &&
			balanceThreshold.Equal(w.BalanceThreshold) &&
			w.Enabled
	})
}

func matchesCreatedEvent(fundingAmount, balanceThreshold amount.Amount) any {
	return mock.MatchedBy(func(evt faucetwallet.CreatedEvent) bool {
		return evt.NetworkId == validNetworkId &&
			evt.DltAccountId == validDltAccountId &&
			fundingAmount.Equal(evt.FundingAmount) &&
			balanceThreshold.Equal(evt.BalanceThreshold) &&
			evt.Enabled
	})
}

func TestDltIngressCreateFaucetWalletCommandHandler_Handle_Success(t *testing.T) {
	ctx := context.Background()
	handler, faucetWalletRepo, custodyKeyRepo, provider, eventBus := newHandler()

	faucetWalletRepo.On("ExistByNetworkId", ctx, validNetworkId).Return(false, nil)
	provider.On("CreateKey", ctx, mock.AnythingOfType("*custody.CreateKeyRequest")).
		Return(&custody.KeyResponse{DltAccountId: validDltAccountId, ExternalId: validExternalId}, nil)
	custodyKeyRepo.On("Save", ctx, matchesCustodyKey(validDltAccountId)).Return(nil)
	faucetWalletRepo.On("Save", ctx, matchesCreatedWallet(*amount.Ten(), *amount.One())).Return(nil)
	eventBus.On("Publish", ctx, matchesCreatedEvent(*amount.Ten(), *amount.One())).Return(nil)

	cmd := Command{NetworkId: validNetworkId, FundingAmount: *amount.Ten(), BalanceThreshold: *amount.One()}
	resp, err := handler.Handle(ctx, &cmd)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, validNetworkId, resp.NetworkId)
	assert.Equal(t, validDltAccountId, resp.DltAccountId)
	assert.True(t, amount.Ten().Equal(resp.FundingAmount))
	assert.True(t, amount.One().Equal(resp.BalanceThreshold))
	assert.True(t, resp.Enabled)
	faucetWalletRepo.AssertExpectations(t)
	custodyKeyRepo.AssertExpectations(t)
	provider.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestDltIngressCreateFaucetWalletCommandHandler_Handle_UnknownNetwork(t *testing.T) {
	ctx := context.Background()
	handler, faucetWalletRepo, custodyKeyRepo, provider, eventBus := newHandler()

	cmd := Command{NetworkId: "unknown-network", FundingAmount: *amount.Ten(), BalanceThreshold: *amount.One()}
	resp, err := handler.Handle(ctx, &cmd)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	faucetWalletRepo.AssertNotCalled(t, "ExistByNetworkId", mock.Anything, mock.Anything)
	provider.AssertNotCalled(t, "CreateKey", mock.Anything, mock.Anything)
	custodyKeyRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
	eventBus.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
}

func TestDltIngressCreateFaucetWalletCommandHandler_Handle_AlreadyExists(t *testing.T) {
	ctx := context.Background()
	handler, faucetWalletRepo, custodyKeyRepo, provider, eventBus := newHandler()

	faucetWalletRepo.On("ExistByNetworkId", ctx, validNetworkId).Return(true, nil)

	cmd := Command{NetworkId: validNetworkId, FundingAmount: *amount.Ten(), BalanceThreshold: *amount.One()}
	resp, err := handler.Handle(ctx, &cmd)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), domainerrors.ErrorCodeFaucetWalletAlreadyExists)
	provider.AssertNotCalled(t, "CreateKey", mock.Anything, mock.Anything)
	custodyKeyRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
	eventBus.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
}

func TestDltIngressCreateFaucetWalletCommandHandler_Handle_ExistByNetworkIdError(t *testing.T) {
	ctx := context.Background()
	handler, faucetWalletRepo, custodyKeyRepo, provider, eventBus := newHandler()

	faucetWalletRepo.On("ExistByNetworkId", ctx, validNetworkId).Return(false, assert.AnError)

	cmd := Command{NetworkId: validNetworkId, FundingAmount: *amount.Ten(), BalanceThreshold: *amount.One()}
	resp, err := handler.Handle(ctx, &cmd)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	provider.AssertNotCalled(t, "CreateKey", mock.Anything, mock.Anything)
	custodyKeyRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
	eventBus.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
}

func TestDltIngressCreateFaucetWalletCommandHandler_Handle_CustodyProviderError(t *testing.T) {
	ctx := context.Background()
	handler, faucetWalletRepo, custodyKeyRepo, provider, eventBus := newHandler()

	faucetWalletRepo.On("ExistByNetworkId", ctx, validNetworkId).Return(false, nil)
	provider.On("CreateKey", ctx, mock.Anything).
		Return((*custody.KeyResponse)(nil), assert.AnError)

	cmd := Command{NetworkId: validNetworkId, FundingAmount: *amount.Ten(), BalanceThreshold: *amount.One()}
	resp, err := handler.Handle(ctx, &cmd)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	custodyKeyRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
	faucetWalletRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
	eventBus.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
}

func TestDltIngressCreateFaucetWalletCommandHandler_Handle_CustodyKeyRepositoryError(t *testing.T) {
	ctx := context.Background()
	handler, faucetWalletRepo, custodyKeyRepo, provider, eventBus := newHandler()

	faucetWalletRepo.On("ExistByNetworkId", ctx, validNetworkId).Return(false, nil)
	provider.On("CreateKey", ctx, mock.Anything).
		Return(&custody.KeyResponse{DltAccountId: validDltAccountId, ExternalId: validExternalId}, nil)
	custodyKeyRepo.On("Save", ctx, matchesCustodyKey(validDltAccountId)).Return(assert.AnError)

	cmd := Command{NetworkId: validNetworkId, FundingAmount: *amount.Ten(), BalanceThreshold: *amount.One()}
	resp, err := handler.Handle(ctx, &cmd)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	faucetWalletRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
	eventBus.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
}

func TestDltIngressCreateFaucetWalletCommandHandler_Handle_FaucetWalletRepositoryError(t *testing.T) {
	ctx := context.Background()
	handler, faucetWalletRepo, custodyKeyRepo, provider, eventBus := newHandler()

	faucetWalletRepo.On("ExistByNetworkId", ctx, validNetworkId).Return(false, nil)
	provider.On("CreateKey", ctx, mock.Anything).
		Return(&custody.KeyResponse{DltAccountId: validDltAccountId, ExternalId: validExternalId}, nil)
	custodyKeyRepo.On("Save", ctx, matchesCustodyKey(validDltAccountId)).Return(nil)
	faucetWalletRepo.On("Save", ctx, matchesCreatedWallet(*amount.Ten(), *amount.One())).Return(assert.AnError)

	cmd := Command{NetworkId: validNetworkId, FundingAmount: *amount.Ten(), BalanceThreshold: *amount.One()}
	resp, err := handler.Handle(ctx, &cmd)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	eventBus.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
}

func TestDltIngressCreateFaucetWalletCommandHandler_Handle_InvalidBalanceThreshold(t *testing.T) {
	ctx := context.Background()
	handler, faucetWalletRepo, custodyKeyRepo, provider, eventBus := newHandler()

	faucetWalletRepo.On("ExistByNetworkId", ctx, validNetworkId).Return(false, nil)
	provider.On("CreateKey", ctx, mock.Anything).
		Return(&custody.KeyResponse{DltAccountId: validDltAccountId, ExternalId: validExternalId}, nil)
	custodyKeyRepo.On("Save", ctx, matchesCustodyKey(validDltAccountId)).Return(nil)

	cmd := Command{NetworkId: validNetworkId, FundingAmount: *amount.Ten(), BalanceThreshold: *amount.Ten()}
	resp, err := handler.Handle(ctx, &cmd)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), domainerrors.ErrorCodeInvalidBalanceThreshold)
	faucetWalletRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
	eventBus.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
}

func TestDltIngressCreateFaucetWalletCommandHandler_Handle_EventBusError(t *testing.T) {
	ctx := context.Background()
	handler, faucetWalletRepo, custodyKeyRepo, provider, eventBus := newHandler()

	faucetWalletRepo.On("ExistByNetworkId", ctx, validNetworkId).Return(false, nil)
	provider.On("CreateKey", ctx, mock.Anything).
		Return(&custody.KeyResponse{DltAccountId: validDltAccountId, ExternalId: validExternalId}, nil)
	custodyKeyRepo.On("Save", ctx, matchesCustodyKey(validDltAccountId)).Return(nil)
	faucetWalletRepo.On("Save", ctx, matchesCreatedWallet(*amount.Ten(), *amount.One())).Return(nil)
	eventBus.On("Publish", ctx, matchesCreatedEvent(*amount.Ten(), *amount.One())).Return(assert.AnError)

	cmd := Command{NetworkId: validNetworkId, FundingAmount: *amount.Ten(), BalanceThreshold: *amount.One()}
	resp, err := handler.Handle(ctx, &cmd)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	faucetWalletRepo.AssertExpectations(t)
	custodyKeyRepo.AssertExpectations(t)
	provider.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}
