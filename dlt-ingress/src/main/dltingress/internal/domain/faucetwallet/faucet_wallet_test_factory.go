package faucetwallet

import (
	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/base"
)

type FaucetWalletConfigurator func(*FaucetWallet)

type FaucetWalletTestFactory struct{}

func NewFaucetWalletTestFactory() *FaucetWalletTestFactory {
	return &FaucetWalletTestFactory{}
}

func (f *FaucetWalletTestFactory) CreateEntity(config ...FaucetWalletConfigurator) *FaucetWallet {
	wallet := &FaucetWallet{
		Entity: base.Entity{
			Id: uuid.New(),
		},
		NetworkId:        "ethereum-sepolia",
		CustodyKeyId:     uuid.New(),
		FundingAmount:    *amount.Ten(),
		BalanceThreshold: *amount.One(),
		Enabled:          true,
	}
	for _, c := range config {
		c(wallet)
	}
	return wallet
}
