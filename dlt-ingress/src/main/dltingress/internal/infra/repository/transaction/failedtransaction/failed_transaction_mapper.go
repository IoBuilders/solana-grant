package failedtransactionrepo

import (
	"dlt-ingress/src/main/dltingress/internal/domain/transaction/failedtransaction"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/base"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/repository/base"
)

func ToDomain(model FailedTransaction) failedtransaction.FailedTransaction {
	return failedtransaction.FailedTransaction{
		Entity: base.Entity{
			Id:        model.Id,
			CreatedAt: model.CreatedAt,
			UpdatedAt: model.UpdatedAt,
		},
		TxId:         model.TxId,
		NetworkId:    model.NetworkId,
		Status:       model.Status,
		ErrorDetails: model.ErrorDetails,
	}
}

func ToDomainList(models []FailedTransaction) []failedtransaction.FailedTransaction {
	domains := make([]failedtransaction.FailedTransaction, len(models))
	for i, model := range models {
		domains[i] = ToDomain(model)
	}
	return domains
}

func FromDomain(domain failedtransaction.FailedTransaction) FailedTransaction {
	return FailedTransaction{
		Model: baserepo.Model{
			Id:        domain.Id,
			CreatedAt: domain.CreatedAt,
			UpdatedAt: domain.UpdatedAt,
		},
		TxId:         domain.TxId,
		NetworkId:    domain.NetworkId,
		Status:       domain.Status,
		ErrorDetails: domain.ErrorDetails,
	}
}
