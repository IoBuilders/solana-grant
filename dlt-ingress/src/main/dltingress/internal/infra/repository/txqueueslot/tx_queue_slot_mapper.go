package txqueueslotrepo

import "dlt-ingress/src/main/dltingress/internal/domain/txqueueslot"

func ToDomain(model TxQueueSlot) txqueueslot.TxQueueSlot {
	return txqueueslot.TxQueueSlot{
		Id:        model.Id,
		CreatedAt: model.CreatedAt,
		NetworkId: model.NetworkId,
		ExpiresAt: model.ExpiresAt,
	}
}

func ToDomainList(models []TxQueueSlot) []txqueueslot.TxQueueSlot {
	domains := make([]txqueueslot.TxQueueSlot, len(models))
	for i, model := range models {
		domains[i] = ToDomain(model)
	}
	return domains
}

func FromDomain(domain txqueueslot.TxQueueSlot) TxQueueSlot {
	return TxQueueSlot{
		Id:        domain.Id,
		CreatedAt: domain.CreatedAt,
		NetworkId: domain.NetworkId,
		ExpiresAt: domain.ExpiresAt,
	}
}
