package retrytransaction

import (
	"dlt-ingress/src/main/dltingress/domain/transaction"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
)

type Command struct {
	TransactionHash      string
	GasLimit             *amount.Amount
	GasPrice             *amount.Amount
	MaxPriorityFeePerGas *amount.Amount
	UseMultiplier        bool
	Nonce                *amount.Amount
}

type Response struct {
	transaction.RetriedEvent
}
