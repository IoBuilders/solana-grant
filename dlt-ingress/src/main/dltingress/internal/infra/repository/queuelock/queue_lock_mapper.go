package queuelockrepo

import (
	"dlt-ingress/src/main/dltingress/internal/domain/queuelock"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/base"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/repository/base"
)

func ToDomain(model QueueLock) queuelock.QueueLock {
	return queuelock.QueueLock{
		Entity: base.Entity{
			Id:        model.Id,
			CreatedAt: model.CreatedAt,
			UpdatedAt: model.UpdatedAt,
		},
		NetworkId: model.NetworkId,
	}
}

func ToDomainList(models []QueueLock) []queuelock.QueueLock {
	domains := make([]queuelock.QueueLock, len(models))
	for i, model := range models {
		domains[i] = ToDomain(model)
	}
	return domains
}

func FromDomain(domain queuelock.QueueLock) QueueLock {
	return QueueLock{
		Model: baserepo.Model{
			Id:        domain.Id,
			CreatedAt: domain.CreatedAt,
			UpdatedAt: domain.UpdatedAt,
		},
		NetworkId: domain.NetworkId,
	}
}
