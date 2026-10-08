package evmtransactionrepo

import (
	"dlt-ingress/src/main/dltingress/internal/domain/transaction/evmtransaction"
	"dlt-ingress/src/main/dltingress/internal/infra/repository/transaction"
)

func ToDomain(model EvmTransaction) evmtransaction.EvmTransaction {
	return evmtransaction.EvmTransaction{
		Transaction:          transactionrepo.ToDomain(model.Transaction),
		Dlt:                  model.Dlt,
		FromAddress:          model.FromAddress,
		ToAddress:            model.ToAddress,
		Nonce:                model.Nonce,
		Value:                model.Value,
		TransactionType:      model.TransactionType,
		GasLimit:             model.GasLimit,
		GasPrice:             model.GasPrice,
		MaxPriorityFeePerGas: model.MaxPriorityFeePerGas,
		MaxFeePerGas:         model.MaxFeePerGas,
		Data:                 model.Data,
	}
}

func ToDomainList(models []EvmTransaction) []evmtransaction.EvmTransaction {
	domains := make([]evmtransaction.EvmTransaction, len(models))
	for i, model := range models {
		domains[i] = ToDomain(model)
	}
	return domains
}

func FromDomain(domain evmtransaction.EvmTransaction) EvmTransaction {
	return EvmTransaction{
		Transaction:          transactionrepo.FromDomain(domain.Transaction),
		Dlt:                  domain.Dlt,
		FromAddress:          domain.FromAddress,
		ToAddress:            domain.ToAddress,
		Nonce:                domain.Nonce,
		Value:                domain.Value,
		TransactionType:      domain.TransactionType,
		GasLimit:             domain.GasLimit,
		GasPrice:             domain.GasPrice,
		MaxPriorityFeePerGas: domain.MaxPriorityFeePerGas,
		MaxFeePerGas:         domain.MaxFeePerGas,
		Data:                 domain.Data,
	}
}
