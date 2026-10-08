package mapper

import (
	"dlt-ingress/src/main/dltingress/internal/app/query"
	"dlt-ingress/src/main/dltingress/port/crossquery"
)

func FromDomain(transaction query.Transaction) *crossquery.Transaction {
	tx := &crossquery.Transaction{
		TxId:       transaction.TxId,
		NetworkId:  transaction.NetworkId,
		NetworkUrl: transaction.NetworkUrl,
		Dlt:        transaction.Dlt,
	}
	if transaction.Evm != nil {
		tx.Evm = &crossquery.EvmTransaction{
			FromAddress:          transaction.Evm.FromAddress,
			ToAddress:            transaction.Evm.ToAddress,
			Nonce:                transaction.Evm.Nonce,
			Value:                transaction.Evm.Value,
			TransactionType:      transaction.Evm.TransactionType,
			GasLimit:             transaction.Evm.GasLimit,
			GasPrice:             transaction.Evm.GasPrice,
			MaxPriorityFeePerGas: transaction.Evm.MaxPriorityFeePerGas,
			MaxFeePerGas:         transaction.Evm.MaxFeePerGas,
			Data:                 transaction.Evm.Data,
		}
	}
	return tx
}
