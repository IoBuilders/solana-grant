package faucetwallet

import (
	"strings"
	"testing"

	"dlt-ingress/src/main/dltingress/internal/domain/domainerrors"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	coredomainerrors "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
)

func TestDltIngressNewFaucetWallet_Success(t *testing.T) {
	custodyKeyId := uuid.New()

	wallet, err := NewFaucetWallet("ethereum-sepolia", custodyKeyId, *amount.Ten(), *amount.One())

	assert.Nil(t, err)
	assert.NotNil(t, wallet)
	assert.Equal(t, "ethereum-sepolia", wallet.NetworkId)
	assert.Equal(t, custodyKeyId, wallet.CustodyKeyId)
	assert.True(t, amount.Ten().Equal(wallet.FundingAmount))
	assert.True(t, amount.One().Equal(wallet.BalanceThreshold))
	assert.True(t, wallet.Enabled)
}

func TestDltIngressNewFaucetWallet_EmptyNetworkId(t *testing.T) {
	wallet, err := NewFaucetWallet("", uuid.New(), *amount.Ten(), *amount.One())

	assert.Nil(t, wallet)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), coredomainerrors.ErrorCodeValidation)
}

func TestDltIngressNewFaucetWallet_NetworkIdTooLong(t *testing.T) {
	tooLong := strings.Repeat("a", 256)

	wallet, err := NewFaucetWallet(tooLong, uuid.New(), *amount.Ten(), *amount.One())

	assert.Nil(t, wallet)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), coredomainerrors.ErrorCodeValidation)
}

func TestDltIngressNewFaucetWallet_ZeroFundingAmount(t *testing.T) {
	wallet, err := NewFaucetWallet("ethereum-sepolia", uuid.New(), *amount.Zero(), *amount.Zero())

	assert.Nil(t, wallet)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), domainerrors.ErrorCodeInvalidFundingAmount)
}

func TestDltIngressNewFaucetWallet_NegativeBalanceThreshold(t *testing.T) {
	negative, err := amount.NewFromString("-1")
	assert.Nil(t, err)

	wallet, err := NewFaucetWallet("ethereum-sepolia", uuid.New(), *amount.Ten(), *negative)

	assert.Nil(t, wallet)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), domainerrors.ErrorCodeInvalidBalanceThreshold)
}

func TestDltIngressNewFaucetWallet_BalanceThresholdNotLowerThanFundingAmount(t *testing.T) {
	wallet, err := NewFaucetWallet("ethereum-sepolia", uuid.New(), *amount.Ten(), *amount.Ten())

	assert.Nil(t, wallet)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), domainerrors.ErrorCodeInvalidBalanceThreshold)
}

func TestDltIngressFaucetWallet_UpdateFundingAmount_Success(t *testing.T) {
	factory := NewFaucetWalletTestFactory()
	wallet := factory.CreateEntity()

	err := wallet.UpdateFundingAmount(*amount.Hundred())

	assert.Nil(t, err)
	assert.True(t, amount.Hundred().Equal(wallet.FundingAmount))
}

func TestDltIngressFaucetWallet_UpdateFundingAmount_Invalid(t *testing.T) {
	factory := NewFaucetWalletTestFactory()
	wallet := factory.CreateEntity()

	err := wallet.UpdateFundingAmount(*amount.Zero())

	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), domainerrors.ErrorCodeInvalidFundingAmount)
	assert.True(t, amount.Ten().Equal(wallet.FundingAmount))
}

func TestDltIngressFaucetWallet_UpdateBalanceThreshold_Success(t *testing.T) {
	factory := NewFaucetWalletTestFactory()
	wallet := factory.CreateEntity()

	err := wallet.UpdateBalanceThreshold(*amount.Zero())

	assert.Nil(t, err)
	assert.True(t, amount.Zero().Equal(wallet.BalanceThreshold))
}

func TestDltIngressFaucetWallet_UpdateBalanceThreshold_Invalid(t *testing.T) {
	factory := NewFaucetWalletTestFactory()
	wallet := factory.CreateEntity()

	err := wallet.UpdateBalanceThreshold(*amount.Ten())

	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), domainerrors.ErrorCodeInvalidBalanceThreshold)
	assert.True(t, amount.One().Equal(wallet.BalanceThreshold))
}

func TestDltIngressFaucetWallet_SetEnabled(t *testing.T) {
	factory := NewFaucetWalletTestFactory()
	wallet := factory.CreateEntity()

	wallet.SetEnabled(false)
	assert.False(t, wallet.Enabled)

	wallet.SetEnabled(true)
	assert.True(t, wallet.Enabled)
}
