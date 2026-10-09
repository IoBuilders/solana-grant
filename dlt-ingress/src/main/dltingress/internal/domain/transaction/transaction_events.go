package transaction

import (
	"dlt-ingress/src/main/dltingress/internal/domain/common"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"
)

type TransactionSentEvent struct {
	event.BaseEvent
	TxId      string `json:"txId" required:"true" description:"Deterministic transaction identifier, known before the transaction is submitted."`
	NetworkId string `json:"networkId" required:"true" description:"Network the transaction was submitted to."`
	Dlt       string `json:"dlt" required:"true" enum:"EVM,SVM" description:"Discriminates which of the two field sets is present."`
	*EvmTransactionEventModel
	*SvmTransactionEventModel
}

type EvmTransactionEventModel struct {
	FromAddress          string         `json:"fromAddress" description:"EVM variant."`
	ToAddress            string         `json:"toAddress" description:"EVM variant."`
	Nonce                amount.Amount  `json:"nonce" description:"EVM variant."`
	Value                *amount.Amount `json:"value" description:"EVM variant."`
	Data                 string         `json:"data" description:"EVM variant. ABI-encoded call data."`
	TransactionType      uint           `json:"transactionType" enum:"0,2" description:"EVM variant. 0 = legacy, 2 = EIP-1559."`
	GasLimit             amount.Amount  `json:"gasLimit" description:"EVM variant."`
	GasPrice             amount.Amount  `json:"gasPrice" description:"EVM variant."`
	MaxPriorityFeePerGas amount.Amount  `json:"maxPriorityFeePerGas" description:"EVM variant."`
	MaxFeePerGas         amount.Amount  `json:"maxFeePerGas" description:"EVM variant."`
}

func NewEvmTransactionSentEvent(txId, networkId string, evmTransactionEventModel *EvmTransactionEventModel) TransactionSentEvent {
	return TransactionSentEvent{
		BaseEvent:                *event.NewBaseEvent(),
		TxId:                     txId,
		NetworkId:                networkId,
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
	FeePayer              string         `json:"feePayer" description:"SVM variant."`
	RecentBlockhash       string         `json:"recentBlockhash" description:"SVM variant."`
	SerializedTransaction string         `json:"serializedTransaction" description:"SVM variant."`
	CuLimit               *amount.Amount `json:"cuLimit" description:"SVM variant."`
	CuPrice               *amount.Amount `json:"cuPrice" description:"SVM variant."`
}

func NewSvmTransactionSentEvent(txId, networkId string, svmTransactionEventModel *SvmTransactionEventModel) TransactionSentEvent {
	return TransactionSentEvent{
		BaseEvent:                *event.NewBaseEvent(),
		TxId:                     txId,
		NetworkId:                networkId,
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
	OldTxId string `json:"oldTxId" required:"true" description:"The superseded transaction."`
	NewTxId string `json:"newTxId" required:"true" description:"The replacement transaction."`
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
