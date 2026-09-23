package evmtransaction

import (
	"dlt-ingress/src/main/dltingress/domain/common"
	"dlt-ingress/src/main/dltingress/domain/domainerrors"
	"dlt-ingress/src/main/dltingress/domain/transaction"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/validate"
)

var validTransactionTypes = map[uint]struct{}{
	0: {},
	2: {},
}

type EvmTransaction struct {
	transaction.Transaction
	Dlt                  common.Dlt     `gorm:"type:varchar(20) not null"`
	FromAddress          string         `gorm:"type:varchar(255) not null"`
	ToAddress            string         `gorm:"type:varchar(255) not null"`
	Nonce                *amount.Amount `gorm:"type:varchar(100) not null"`
	Value                *amount.Amount `gorm:"type:varchar(100) not null"`
	TransactionType      uint           `gorm:"not null"`
	GasLimit             *amount.Amount `gorm:"type:varchar(100) not null"`
	GasPrice             *amount.Amount `gorm:"type:varchar(100)"`
	MaxPriorityFeePerGas *amount.Amount `gorm:"type:varchar(100)"`
	MaxFeePerGas         *amount.Amount `gorm:"type:varchar(100)"`
	Data                 string         `gorm:"type:text"`
}

func NewEvmTransaction(
	txId string,
	networkId string,
	networkUrl string,
	dlt string,
	fromAddress string,
	toAddress string,
	nonce *amount.Amount,
	value *amount.Amount,
	transactionType uint,
	gasLimit *amount.Amount,
	gasPrice *amount.Amount,
	maxPriorityFeePerGas *amount.Amount,
	maxFeePerGas *amount.Amount,
	data string,
) (*EvmTransaction, error) {
	if err := validate.StringMaxLength(txId, "TxId", "EvmTransaction", 255); err != nil {
		return nil, err
	}
	if err := validate.StringMaxLength(networkId, "NetworkId", "EvmTransaction", 255); err != nil {
		return nil, err
	}
	if err := validate.StringMaxLength(networkUrl, "NetworkUrl", "EvmTransaction", 255); err != nil {
		return nil, err
	}
	parsedDlt, err := common.ParseDlt(dlt)
	if err != nil {
		return nil, err
	}
	if err := validate.StringMaxLength(fromAddress, "FromAddress", "EvmTransaction", 255); err != nil {
		return nil, err
	}
	if err := validate.StringMaxLength(toAddress, "ToAddress", "EvmTransaction", 255); err != nil {
		return nil, err
	}
	if _, ok := validTransactionTypes[transactionType]; !ok {
		return nil, domainerrors.NewInvalidTransactionTypeDomainError(transactionType)
	}

	return &EvmTransaction{
		Transaction: transaction.Transaction{
			TxId:       txId,
			NetworkId:  networkId,
			NetworkUrl: networkUrl,
		},
		Dlt:                  parsedDlt,
		FromAddress:          fromAddress,
		ToAddress:            toAddress,
		Nonce:                nonce,
		Value:                value,
		TransactionType:      transactionType,
		GasLimit:             gasLimit,
		GasPrice:             gasPrice,
		MaxPriorityFeePerGas: maxPriorityFeePerGas,
		MaxFeePerGas:         maxFeePerGas,
		Data:                 data,
	}, nil
}

func (tx *EvmTransaction) Clone() *EvmTransaction {
	if tx == nil {
		return nil
	}

	return &EvmTransaction{
		Transaction:          tx.Transaction,
		Dlt:                  tx.Dlt,
		FromAddress:          tx.FromAddress,
		ToAddress:            tx.ToAddress,
		Nonce:                cloneAmount(tx.Nonce),
		Value:                cloneAmount(tx.Value),
		TransactionType:      tx.TransactionType,
		GasLimit:             cloneAmount(tx.GasLimit),
		GasPrice:             cloneAmount(tx.GasPrice),
		MaxPriorityFeePerGas: cloneAmount(tx.MaxPriorityFeePerGas),
		MaxFeePerGas:         cloneAmount(tx.MaxFeePerGas),
		Data:                 tx.Data,
	}
}

func cloneAmount(a *amount.Amount) *amount.Amount {
	if a == nil {
		return nil
	}
	b, _ := amount.New(a.RawValue(), a.Decimals())
	return b
}
