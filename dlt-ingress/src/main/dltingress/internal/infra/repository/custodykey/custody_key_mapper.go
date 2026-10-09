package custodykeyrepo

import (
	"dlt-ingress/src/main/dltingress/internal/domain/custodykey"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/base"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/repository/base"
)

func ToDomain(model CustodyKey) custodykey.CustodyKey {
	return custodykey.CustodyKey{
		Entity: base.Entity{
			Id:        model.Id,
			CreatedAt: model.CreatedAt,
			UpdatedAt: model.UpdatedAt,
		},
		KeyType:         model.KeyType,
		Status:          model.Status,
		DltAccountId:    model.DltAccountId,
		Dlt:             model.Dlt,
		ExternalId:      model.ExternalId,
		CustodyProvider: model.CustodyProvider,
	}
}

func ToDomainList(models []CustodyKey) []custodykey.CustodyKey {
	domains := make([]custodykey.CustodyKey, len(models))
	for i, dltAccountId := range models {
		domains[i] = ToDomain(dltAccountId)
	}
	return domains
}

func ToDomainPointerList(models []*CustodyKey) []*custodykey.CustodyKey {
	domains := make([]*custodykey.CustodyKey, len(models))
	for i, dltAccountId := range models {
		domains[i] = new(ToDomain(*dltAccountId))
	}
	return domains
}

func FromDomain(domain custodykey.CustodyKey) CustodyKey {
	return CustodyKey{
		Model: baserepo.Model{
			Id:        domain.Id,
			CreatedAt: domain.CreatedAt,
			UpdatedAt: domain.UpdatedAt,
		},
		KeyType:         domain.KeyType,
		Status:          domain.Status,
		DltAccountId:    domain.DltAccountId,
		Dlt:             domain.Dlt,
		ExternalId:      domain.ExternalId,
		CustodyProvider: domain.CustodyProvider,
	}
}
