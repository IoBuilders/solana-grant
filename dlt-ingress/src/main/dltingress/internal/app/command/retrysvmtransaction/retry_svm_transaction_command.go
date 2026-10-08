package retrysvmtransaction

import (
	"dlt-ingress/src/main/dltingress/internal/domain/transaction"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
)

type Command struct {
	TransactionHash string
	CuLimit         *amount.Amount
	CuPrice         *amount.Amount
	UseMultiplier   bool
}

type Response struct {
	transaction.RetriedEvent
}
