package model

import "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"

// FaucetWalletModel represents a FaucetWallet.
type FaucetWalletModel struct {
	Id           string `json:"id" example:"9e2e6b8a-6b39-4d3a-9f0a-1a2b3c4d5e6f"`
	NetworkId    string `json:"networkId" example:"ethereum-sepolia"`
	DltAccountId string `json:"dltAccountId" example:"0x8f6c...9a2b"`
	// FundingAmount is expressed in the network's smallest native unit (e.g. wei for EVM networks).
	FundingAmount amount.Amount `json:"fundingAmount" swaggertype:"string" format:"decimal" example:"1000000000000000000"`
	// BalanceThreshold is expressed in the network's smallest native unit (e.g. wei for EVM networks).
	BalanceThreshold amount.Amount `json:"balanceThreshold" swaggertype:"string" format:"decimal" example:"100000000000000000"`
	Enabled          bool          `json:"enabled" example:"true"`
} // @name FaucetWallet
