package createfaucetwallet

import "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"

type Request struct {
	// FundingAmount is expressed in the network's smallest native unit (e.g. wei for EVM networks).
	FundingAmount amount.Amount `json:"fundingAmount" binding:"required" example:"1000000000000000000"`
	// BalanceThreshold is expressed in the network's smallest native unit (e.g. wei for EVM networks).
	BalanceThreshold amount.Amount `json:"balanceThreshold" binding:"required" example:"100000000000000000"`
} // @name CreateFaucetWalletRequest
