package transactionrepo

import (
	"dlt-ingress/src/main/dltingress/internal/domain/transaction"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/base"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/repository/base"
)

func ToDomain(model Transaction) transaction.Transaction {
	return transaction.Transaction{
		Entity: base.Entity{
			Id:        model.Id,
			CreatedAt: model.CreatedAt,
			UpdatedAt: model.UpdatedAt,
		},
		TxId:         model.TxId,
		OriginalTxId: model.OriginalTxId,
		NetworkId:    model.NetworkId,
		NetworkUrl:   model.NetworkUrl,
	}
}

func ToDomainList(models []Transaction) []transaction.Transaction {
	domains := make([]transaction.Transaction, len(models))
	for i, model := range models {
		domains[i] = ToDomain(model)
	}
	return domains
}

func FromDomain(domain transaction.Transaction) Transaction {
	return Transaction{
		Model: baserepo.Model{
			Id:        domain.Id,
			CreatedAt: domain.CreatedAt,
			UpdatedAt: domain.UpdatedAt,
		},
		TxId:         domain.TxId,
		OriginalTxId: domain.OriginalTxId,
		NetworkId:    domain.NetworkId,
		NetworkUrl:   domain.NetworkUrl,
	}
}
