package signandsendevents

import (
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"
)

type TransactionSentCrossEvent struct {
	event.BaseEvent
	TxId string
	Dlt  string
	*EvmTransactionCrossEventModel
	*SvmTransactionCrossEventModel
}

type EvmTransactionCrossEventModel struct {
	FromAddress          string
	ToAddress            string
	Nonce                amount.Amount
	Value                *amount.Amount
	Data                 string
	TransactionType      uint
	GasLimit             amount.Amount
	GasPrice             amount.Amount
	MaxPriorityFeePerGas amount.Amount
	MaxFeePerGas         amount.Amount
}

type SvmTransactionCrossEventModel struct {
	FeePayer              string
	RecentBlockhash       string
	SerializedTransaction string
	CuLimit               *amount.Amount
	CuPrice               *amount.Amount
}
