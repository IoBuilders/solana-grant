package evmtransaction

import (
	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction"
	"dlt-ingress/src/main/dltingress/internal/infra/portcommon"

	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/base"
)

type EvmTransactionConfigurator func(*EvmTransaction)

type EvmTransactionTestFactory struct{}

func NewEvmTransactionTestFactory() *EvmTransactionTestFactory {
	return &EvmTransactionTestFactory{}
}

func (f *EvmTransactionTestFactory) CreateEntity(config ...EvmTransactionConfigurator) *EvmTransaction {
	nonce, _ := amount.NewFromString("1")
	value, _ := amount.NewFromString("0")
	gasLimit, _ := amount.NewFromString("21000")
	gasPrice, _ := amount.NewFromString("1000000000")

	entity := &EvmTransaction{
		Transaction: transaction.Transaction{
			Entity: base.Entity{
				Id: uuid.New(),
			},
			TxId:       "0xabc123def456abc123def456abc123def456abc123def456abc123def456abc123",
			NetworkId:  "1",
			NetworkUrl: "http://localhost:8545",
		},
		Dlt:             common.EVM,
		FromAddress:     "0xAbCdEf1234567890AbCdEf1234567890AbCdEf12",
		ToAddress:       "0x1234567890AbCdEf1234567890AbCdEf12345678",
		Nonce:           nonce,
		Value:           value,
		TransactionType: portcommon.TransactionTypeLegacy,
		GasLimit:        gasLimit,
		GasPrice:        gasPrice,
		Data:            "0xa9059cbb000000000000000000000000",
	}
	for _, c := range config {
		c(entity)
	}
	return entity
}

func (f *EvmTransactionTestFactory) CreateEntityDynamicFee(config ...EvmTransactionConfigurator) *EvmTransaction {
	nonce, _ := amount.NewFromString("1")
	value, _ := amount.NewFromString("0")
	gasLimit, _ := amount.NewFromString("21000")
	maxPriorityFeePerGas, _ := amount.NewFromString("1000000000")
	MaxFeePerGas, _ := amount.NewFromString("1000000000")

	entity := &EvmTransaction{
		Transaction: transaction.Transaction{
			Entity: base.Entity{
				Id: uuid.New(),
			},
			TxId:       "0xabc123def456abc123def456abc123def456abc123def456abc123def456abc123",
			NetworkId:  "1",
			NetworkUrl: "http://localhost:8545",
		},
		Dlt:                  common.EVM,
		FromAddress:          "0xAbCdEf1234567890AbCdEf1234567890AbCdEf12",
		ToAddress:            "0x1234567890AbCdEf1234567890AbCdEf12345678",
		Nonce:                nonce,
		Value:                value,
		TransactionType:      portcommon.TransactionTypeDynamicFee,
		GasLimit:             gasLimit,
		MaxPriorityFeePerGas: maxPriorityFeePerGas,
		MaxFeePerGas:         MaxFeePerGas,
		Data:                 "0xa9059cbb000000000000000000000000",
	}
	for _, c := range config {
		c(entity)
	}
	return entity
}
