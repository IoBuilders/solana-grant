package faucetwallet

import (
	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"
)

type CreatedEvent struct {
	event.BaseEvent
	FaucetWalletId   uuid.UUID     `json:"faucetWalletId" required:"true" description:"Identity of the faucet wallet."`
	NetworkId        string        `json:"networkId" required:"true" maxLength:"255" description:"Network the wallet funds accounts on. One faucet wallet per network."`
	CustodyKeyId     uuid.UUID     `json:"custodyKeyId" required:"true" description:"The custody key provisioned for this wallet."`
	DltAccountId     string        `json:"dltAccountId" required:"true" description:"Address of the account that funds the top-ups."`
	FundingAmount    amount.Amount `json:"fundingAmount" required:"true" description:"Amount transferred on each top-up, in the network's smallest native unit (e.g. wei for EVM networks)."`
	BalanceThreshold amount.Amount `json:"balanceThreshold" required:"true" description:"Balance below which a top-up is triggered, in the network's smallest native unit. Always lower than fundingAmount."`
	Enabled          bool          `json:"enabled" required:"true" description:"Whether top-ups are performed for this network. Always true on creation."`
}

type UpdatedEvent struct {
	event.BaseEvent
	FaucetWalletId   uuid.UUID     `json:"faucetWalletId" required:"true" description:"Identity of the faucet wallet."`
	NetworkId        string        `json:"networkId" required:"true" maxLength:"255" description:"Network the wallet funds accounts on. One faucet wallet per network."`
	FundingAmount    amount.Amount `json:"fundingAmount" required:"true" description:"Amount transferred on each top-up, in the network's smallest native unit (e.g. wei for EVM networks)."`
	BalanceThreshold amount.Amount `json:"balanceThreshold" required:"true" description:"Balance below which a top-up is triggered, in the network's smallest native unit. Always lower than fundingAmount."`
	Enabled          bool          `json:"enabled" required:"true" description:"Whether top-ups are performed for this network."`
}
