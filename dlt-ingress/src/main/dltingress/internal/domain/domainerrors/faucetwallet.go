package domainerrors

import (
	"fmt"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
)

const (
	ErrorCodeInvalidFundingAmount            coreerror.ErrorCode = "INVALID_FUNDING_AMOUNT"
	ErrorCodeInvalidBalanceThreshold         coreerror.ErrorCode = "INVALID_BALANCE_THRESHOLD"
	ErrorCodeFaucetWalletNotFoundByNetworkId coreerror.ErrorCode = "FAUCET_WALLET_NOT_FOUND_BY_NETWORK_ID"
	ErrorCodeFaucetWalletAlreadyExists       coreerror.ErrorCode = "FAUCET_WALLET_ALREADY_EXISTS"
	ErrorCodeFaucetWalletInsufficientFunds   coreerror.ErrorCode = "FAUCET_WALLET_INSUFFICIENT_FUNDS"
)

func NewInvalidFundingAmountDomainError(fundingAmount amount.Amount) error {
	return coreerror.NewValidationDomainError(
		ErrorCodeInvalidFundingAmount,
		fmt.Sprintf("invalid funding amount: %s (must be greater than zero)", fundingAmount.String()),
	)
}

func NewInvalidBalanceThresholdDomainError(balanceThreshold amount.Amount) error {
	return coreerror.NewValidationDomainError(
		ErrorCodeInvalidBalanceThreshold,
		fmt.Sprintf("invalid balance threshold: %s (must be non-negative and lower than the funding amount)", balanceThreshold.String()),
	)
}

func NewFaucetWalletNotFoundByNetworkIdDomainError(networkId string) error {
	return coreerror.NewNotFoundDomainError(
		ErrorCodeFaucetWalletNotFoundByNetworkId,
		fmt.Sprintf("faucet wallet not found for network ID: %s", networkId),
	)
}

func NewFaucetWalletAlreadyExistsDomainError(networkId string) error {
	return coreerror.NewValidationDomainError(
		ErrorCodeFaucetWalletAlreadyExists,
		fmt.Sprintf("a faucet wallet already exists for network ID: %s", networkId),
	)
}

func NewFaucetWalletInsufficientFundsDomainError(networkId string) error {
	return coreerror.NewValidationDomainError(
		ErrorCodeFaucetWalletInsufficientFunds,
		fmt.Sprintf("faucet wallet has insufficient funds for network ID: %s", networkId),
	)
}
