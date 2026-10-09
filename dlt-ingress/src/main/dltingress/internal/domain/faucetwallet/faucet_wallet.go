package faucetwallet

import (
	"dlt-ingress/src/main/dltingress/internal/domain/domainerrors"

	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/base"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/validate"
)

type FaucetWallet struct {
	base.Entity
	NetworkId    string
	CustodyKeyId uuid.UUID
	// FundingAmount is expressed in the network's smallest native unit (e.g. wei for EVM networks).
	FundingAmount amount.Amount
	// BalanceThreshold is expressed in the network's smallest native unit (e.g. wei for EVM networks).
	BalanceThreshold amount.Amount
	Enabled          bool
}

// NewFaucetWallet creates a FaucetWallet
func NewFaucetWallet(networkId string, custodyKeyId uuid.UUID, fundingAmount, balanceThreshold amount.Amount) (*FaucetWallet, error) {
	if err := validate.StringMaxLength(networkId, "NetworkId", "FaucetWallet", 255); err != nil {
		return nil, err
	}
	if err := validateFundingAmount(fundingAmount); err != nil {
		return nil, err
	}
	if err := validateBalanceThreshold(balanceThreshold, fundingAmount); err != nil {
		return nil, err
	}

	return &FaucetWallet{
		NetworkId:        networkId,
		CustodyKeyId:     custodyKeyId,
		FundingAmount:    fundingAmount,
		BalanceThreshold: balanceThreshold,
		Enabled:          true,
	}, nil
}

func (w *FaucetWallet) UpdateFundingAmount(fundingAmount amount.Amount) error {
	if err := validateFundingAmount(fundingAmount); err != nil {
		return err
	}
	w.FundingAmount = fundingAmount
	return nil
}

func (w *FaucetWallet) UpdateBalanceThreshold(balanceThreshold amount.Amount) error {
	if err := validateBalanceThreshold(balanceThreshold, w.FundingAmount); err != nil {
		return err
	}
	w.BalanceThreshold = balanceThreshold
	return nil
}

func (w *FaucetWallet) SetEnabled(enabled bool) {
	w.Enabled = enabled
}

func validateFundingAmount(fundingAmount amount.Amount) error {
	if !fundingAmount.GreaterThan(*amount.Zero()) {
		return domainerrors.NewInvalidFundingAmountDomainError(fundingAmount)
	}
	return nil
}

func validateBalanceThreshold(balanceThreshold, fundingAmount amount.Amount) error {
	if balanceThreshold.LessThan(*amount.Zero()) || !balanceThreshold.LessThan(fundingAmount) {
		return domainerrors.NewInvalidBalanceThresholdDomainError(balanceThreshold)
	}
	return nil
}
