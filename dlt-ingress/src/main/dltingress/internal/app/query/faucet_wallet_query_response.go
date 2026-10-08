package query

import (
	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
)

type FaucetWallet struct {
	FaucetWalletId   uuid.UUID
	NetworkId        string
	DltAccountId     string
	FundingAmount    amount.Amount
	BalanceThreshold amount.Amount
	Balance          amount.Amount
	Enabled          bool
}
