package transaction

import (
	"dlt-ingress/src/main/dltingress/domain/common"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"
)

type TransactionSentEvent struct {
	event.BaseEvent
	TxId string
	Dlt  string
	*EvmTransactionEventModel
	*SvmTransactionEventModel
}

type EvmTransactionEventModel struct {
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

func NewEvmTransactionSentEvent(txId string, evmTransactionEventModel *EvmTransactionEventModel) *TransactionSentEvent {
	return &TransactionSentEvent{
		BaseEvent:                *event.NewBaseEvent(),
		TxId:                     txId,
		Dlt:                      string(common.EVM),
		EvmTransactionEventModel: evmTransactionEventModel,
	}
}

func NewEvmTransactionEventModel(
	fromAddress string,
	toAddress string,
	nonce amount.Amount,
	value *amount.Amount,
	Data string,
	transactionType uint,
	gasLimit amount.Amount,
	gasPrice amount.Amount,
	maxPriorityFeePerGas amount.Amount,
	maxFeePerGas amount.Amount,
) *EvmTransactionEventModel {
	return &EvmTransactionEventModel{
		FromAddress:          fromAddress,
		ToAddress:            toAddress,
		Nonce:                nonce,
		Value:                value,
		Data:                 Data,
		TransactionType:      transactionType,
		GasLimit:             gasLimit,
		GasPrice:             gasPrice,
		MaxPriorityFeePerGas: maxPriorityFeePerGas,
		MaxFeePerGas:         maxFeePerGas,
	}
}

type SvmTransactionEventModel struct {
	FeePayer              string
	RecentBlockhash       string
	SerializedTransaction string
	CuLimit               *amount.Amount
	CuPrice               *amount.Amount
}

func NewSvmTransactionSentEvent(txId string, svmTransactionEventModel *SvmTransactionEventModel) *TransactionSentEvent {
	return &TransactionSentEvent{
		BaseEvent:                *event.NewBaseEvent(),
		TxId:                     txId,
		Dlt:                      string(common.SVM),
		SvmTransactionEventModel: svmTransactionEventModel,
	}
}

func NewSvmTransactionEventModel(
	feePayer string,
	recentBlockhash string,
	serializedTransaction string,
	cuLimit *amount.Amount,
	cuPrice *amount.Amount,
) *SvmTransactionEventModel {
	return &SvmTransactionEventModel{
		FeePayer:              feePayer,
		RecentBlockhash:       recentBlockhash,
		SerializedTransaction: serializedTransaction,
		CuLimit:               cuLimit,
		CuPrice:               cuPrice,
	}
}

type RetriedEvent struct {
	event.BaseEvent
	OldTxId string
	NewTxId string
}

type TransactionBuiltEvent struct {
	event.BaseEvent
	DltAccountId string
	Payload      string
}

func NewTransactionBuiltEvent(dltAccountId string, payload string) *TransactionBuiltEvent {
	return &TransactionBuiltEvent{
		BaseEvent:    *event.NewBaseEvent(),
		DltAccountId: dltAccountId,
		Payload:      payload,
	}
}
