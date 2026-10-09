package updatefaucetwallet

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/infra/repository/faucetwallet/mock"
	"testing"

	"dlt-ingress/src/main/dltingress/internal/domain/domainerrors"
	"dlt-ingress/src/main/dltingress/internal/domain/faucetwallet"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
	testbus "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/test/core/event"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const validNetworkId = "ethereum-sepolia"

func newHandler() (*CommandHandler, *mockfaucetwalletrepo.Postgres, *testbus.EventBusMock) {
	faucetWalletRepo := new(mockfaucetwalletrepo.Postgres)
	eventBus := new(testbus.EventBusMock)
	return NewCommandHandler(faucetWalletRepo, eventBus), faucetWalletRepo, eventBus
}

func createWallet(t *testing.T) *faucetwallet.FaucetWallet {
	wallet, err := faucetwallet.NewFaucetWallet(validNetworkId, uuid.New(), *amount.Ten(), *amount.One())
	assert.NoError(t, err)
	return wallet
}

func newAmount(t *testing.T, value string) amount.Amount {
	a, err := amount.NewFromString(value)
	assert.NoError(t, err)
	return *a
}

func matchesUpdatedWallet(fundingAmount, balanceThreshold amount.Amount, enabled bool) any {
	return mock.MatchedBy(func(w *faucetwallet.FaucetWallet) bool {
		return fundingAmount.Equal(w.FundingAmount) && balanceThreshold.Equal(w.BalanceThreshold) && w.Enabled == enabled
	})
}

func matchesUpdatedEvent(wallet *faucetwallet.FaucetWallet, fundingAmount, balanceThreshold amount.Amount, enabled bool) any {
	return mock.MatchedBy(func(evt faucetwallet.UpdatedEvent) bool {
		return evt.FaucetWalletId == wallet.Id &&
			evt.NetworkId == wallet.NetworkId &&
			fundingAmount.Equal(evt.FundingAmount) &&
			balanceThreshold.Equal(evt.BalanceThreshold) &&
			evt.Enabled == enabled
	})
}

func TestDltIngressUpdateFaucetWalletCommandHandler_Handle_Success(t *testing.T) {
	ctx := context.Background()
	handler, faucetWalletRepo, eventBus := newHandler()
	wallet := createWallet(t)

	newFundingAmount := newAmount(t, "20")
	newBalanceThreshold := newAmount(t, "2")
	newEnabled := false

	faucetWalletRepo.On("FindByNetworkId", ctx, validNetworkId).Return(wallet, nil)
	faucetWalletRepo.On("Save", ctx, matchesUpdatedWallet(newFundingAmount, newBalanceThreshold, newEnabled)).Return(nil)
	eventBus.On("Publish", ctx, matchesUpdatedEvent(wallet, newFundingAmount, newBalanceThreshold, newEnabled)).Return(nil)

	cmd := Command{NetworkId: validNetworkId, FundingAmount: newFundingAmount, BalanceThreshold: newBalanceThreshold, Enabled: newEnabled}
	resp, err := handler.Handle(ctx, &cmd)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, validNetworkId, resp.NetworkId)
	assert.True(t, newFundingAmount.Equal(resp.FundingAmount))
	assert.True(t, newBalanceThreshold.Equal(resp.BalanceThreshold))
	assert.Equal(t, newEnabled, resp.Enabled)
	faucetWalletRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestDltIngressUpdateFaucetWalletCommandHandler_Handle_NotFound(t *testing.T) {
	ctx := context.Background()
	handler, faucetWalletRepo, eventBus := newHandler()

	faucetWalletRepo.On("FindByNetworkId", ctx, validNetworkId).
		Return(nil, domainerrors.NewFaucetWalletNotFoundByNetworkIdDomainError(validNetworkId))

	cmd := Command{NetworkId: validNetworkId, FundingAmount: *amount.Ten(), BalanceThreshold: *amount.One(), Enabled: true}
	resp, err := handler.Handle(ctx, &cmd)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), domainerrors.ErrorCodeFaucetWalletNotFoundByNetworkId)
	faucetWalletRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
	eventBus.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
}

func TestDltIngressUpdateFaucetWalletCommandHandler_Handle_InvalidFundingAmount(t *testing.T) {
	ctx := context.Background()
	handler, faucetWalletRepo, eventBus := newHandler()
	wallet := createWallet(t)

	faucetWalletRepo.On("FindByNetworkId", ctx, validNetworkId).Return(wallet, nil)

	cmd := Command{NetworkId: validNetworkId, FundingAmount: *amount.Zero(), BalanceThreshold: *amount.One(), Enabled: true}
	resp, err := handler.Handle(ctx, &cmd)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), domainerrors.ErrorCodeInvalidFundingAmount)
	faucetWalletRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
	eventBus.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
}

func TestDltIngressUpdateFaucetWalletCommandHandler_Handle_InvalidBalanceThreshold(t *testing.T) {
	ctx := context.Background()
	handler, faucetWalletRepo, eventBus := newHandler()
	wallet := createWallet(t)

	faucetWalletRepo.On("FindByNetworkId", ctx, validNetworkId).Return(wallet, nil)

	cmd := Command{NetworkId: validNetworkId, FundingAmount: *amount.Ten(), BalanceThreshold: *amount.Ten(), Enabled: true}
	resp, err := handler.Handle(ctx, &cmd)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), domainerrors.ErrorCodeInvalidBalanceThreshold)
	faucetWalletRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
	eventBus.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
}

func TestDltIngressUpdateFaucetWalletCommandHandler_Handle_SaveError(t *testing.T) {
	ctx := context.Background()
	handler, faucetWalletRepo, eventBus := newHandler()
	wallet := createWallet(t)

	newFundingAmount := newAmount(t, "20")
	newBalanceThreshold := newAmount(t, "2")
	newEnabled := false

	faucetWalletRepo.On("FindByNetworkId", ctx, validNetworkId).Return(wallet, nil)
	faucetWalletRepo.On("Save", ctx, matchesUpdatedWallet(newFundingAmount, newBalanceThreshold, newEnabled)).Return(assert.AnError)

	cmd := Command{NetworkId: validNetworkId, FundingAmount: newFundingAmount, BalanceThreshold: newBalanceThreshold, Enabled: newEnabled}
	resp, err := handler.Handle(ctx, &cmd)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	eventBus.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
}

func TestDltIngressUpdateFaucetWalletCommandHandler_Handle_EventBusError(t *testing.T) {
	ctx := context.Background()
	handler, faucetWalletRepo, eventBus := newHandler()
	wallet := createWallet(t)

	newFundingAmount := newAmount(t, "20")
	newBalanceThreshold := newAmount(t, "2")
	newEnabled := false

	faucetWalletRepo.On("FindByNetworkId", ctx, validNetworkId).Return(wallet, nil)
	faucetWalletRepo.On("Save", ctx, matchesUpdatedWallet(newFundingAmount, newBalanceThreshold, newEnabled)).Return(nil)
	eventBus.On("Publish", ctx, matchesUpdatedEvent(wallet, newFundingAmount, newBalanceThreshold, newEnabled)).Return(assert.AnError)

	cmd := Command{NetworkId: validNetworkId, FundingAmount: newFundingAmount, BalanceThreshold: newBalanceThreshold, Enabled: newEnabled}
	resp, err := handler.Handle(ctx, &cmd)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	faucetWalletRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}
