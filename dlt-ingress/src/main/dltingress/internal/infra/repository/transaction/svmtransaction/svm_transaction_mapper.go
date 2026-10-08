package svmtransactionrepo

import (
	"dlt-ingress/src/main/dltingress/internal/domain/transaction/svmtransaction"
	"dlt-ingress/src/main/dltingress/internal/infra/repository/transaction"
)

func ToDomain(model SvmTransaction) svmtransaction.SvmTransaction {
	return svmtransaction.SvmTransaction{
		Transaction:           transactionrepo.ToDomain(model.Transaction),
		Dlt:                   model.Dlt,
		FeePayer:              model.FeePayer,
		RecentBlockhash:       model.RecentBlockhash,
		CuLimit:               model.CuLimit,
		CuPrice:               model.CuPrice,
		SerializedTransaction: model.SerializedTransaction,
	}
}

func ToDomainList(models []SvmTransaction) []svmtransaction.SvmTransaction {
	domains := make([]svmtransaction.SvmTransaction, len(models))
	for i, model := range models {
		domains[i] = ToDomain(model)
	}
	return domains
}

func FromDomain(domain svmtransaction.SvmTransaction) SvmTransaction {
	return SvmTransaction{
		Transaction:           transactionrepo.FromDomain(domain.Transaction),
		Dlt:                   domain.Dlt,
		FeePayer:              domain.FeePayer,
		RecentBlockhash:       domain.RecentBlockhash,
		CuLimit:               domain.CuLimit,
		CuPrice:               domain.CuPrice,
		SerializedTransaction: domain.SerializedTransaction,
	}
}
