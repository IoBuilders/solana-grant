package query

import "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"

type Transaction struct {
	TxId       string
	NetworkId  string
	NetworkUrl string
	Dlt        string
	Evm        *EvmTransaction
}

type EvmTransaction struct {
	FromAddress          string
	ToAddress            string
	Nonce                *amount.Amount
	Value                *amount.Amount
	TransactionType      uint
	GasLimit             *amount.Amount
	GasPrice             *amount.Amount
	MaxPriorityFeePerGas *amount.Amount
	MaxFeePerGas         *amount.Amount
	Data                 string
}
