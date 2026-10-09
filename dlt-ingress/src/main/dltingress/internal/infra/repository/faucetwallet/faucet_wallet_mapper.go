package faucetwalletrepo

import (
	"dlt-ingress/src/main/dltingress/internal/domain/faucetwallet"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/base"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/repository/base"
)

func ToDomain(model FaucetWallet) faucetwallet.FaucetWallet {
	return faucetwallet.FaucetWallet{
		Entity: base.Entity{
			Id:        model.Id,
			CreatedAt: model.CreatedAt,
			UpdatedAt: model.UpdatedAt,
		},
		NetworkId:        model.NetworkId,
		CustodyKeyId:     model.CustodyKeyId,
		FundingAmount:    model.FundingAmount,
		BalanceThreshold: model.BalanceThreshold,
		Enabled:          model.Enabled,
	}
}

func ToDomainList(models []FaucetWallet) []faucetwallet.FaucetWallet {
	domains := make([]faucetwallet.FaucetWallet, len(models))
	for i, dltAccountId := range models {
		domains[i] = ToDomain(dltAccountId)
	}
	return domains
}

func FromDomain(domain faucetwallet.FaucetWallet) FaucetWallet {
	return FaucetWallet{
		Model: baserepo.Model{
			Id:        domain.Id,
			CreatedAt: domain.CreatedAt,
			UpdatedAt: domain.UpdatedAt,
		},
		NetworkId:        domain.NetworkId,
		CustodyKeyId:     domain.CustodyKeyId,
		FundingAmount:    domain.FundingAmount,
		BalanceThreshold: domain.BalanceThreshold,
		Enabled:          domain.Enabled,
	}
}
