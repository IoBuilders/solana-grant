package noncerepo

import (
	"dlt-ingress/src/main/dltingress/internal/domain/nonce"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/base"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/repository/base"
)

func ToDomain(model Nonce) nonce.Nonce {
	return nonce.Nonce{
		Entity: base.Entity{
			Id:        model.Id,
			CreatedAt: model.CreatedAt,
			UpdatedAt: model.UpdatedAt,
		},
		DltAccountId: model.DltAccountId,
		NetworkId:    model.NetworkId,
		Value:        model.Value,
	}
}

func ToDomainList(models []Nonce) []nonce.Nonce {
	domains := make([]nonce.Nonce, len(models))
	for i, model := range models {
		domains[i] = ToDomain(model)
	}
	return domains
}

func FromDomain(domain nonce.Nonce) Nonce {
	return Nonce{
		Model: baserepo.Model{
			Id:        domain.Id,
			CreatedAt: domain.CreatedAt,
			UpdatedAt: domain.UpdatedAt,
		},
		DltAccountId: domain.DltAccountId,
		NetworkId:    domain.NetworkId,
		Value:        domain.Value,
	}
}
