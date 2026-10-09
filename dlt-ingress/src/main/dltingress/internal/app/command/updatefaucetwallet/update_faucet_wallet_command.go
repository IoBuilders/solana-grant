package updatefaucetwallet

import (
	"dlt-ingress/src/main/dltingress/internal/domain/faucetwallet"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
)

type Command struct {
	NetworkId        string
	FundingAmount    amount.Amount
	BalanceThreshold amount.Amount
	Enabled          bool
}

type Response struct {
	faucetwallet.UpdatedEvent
}
