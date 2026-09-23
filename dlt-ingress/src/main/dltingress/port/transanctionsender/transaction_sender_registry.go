package transanctionsender

import (
	"dlt-ingress/src/main/dltingress/domain/common"
	"fmt"
)

type Registry struct {
	senders map[common.Dlt]Port
}

func NewRegistry() *Registry {
	return &Registry{
		senders: make(map[common.Dlt]Port),
	}
}

func (r *Registry) Register(dlt common.Dlt, sender Port) {
	r.senders[dlt] = sender
}

func (r *Registry) GetTransactionSender(dlt common.Dlt) (Port, error) {
	sender, found := r.senders[dlt]
	if !found {
		return nil, fmt.Errorf("no transaction sender configured for dlt %s", dlt)
	}
	return sender, nil
}
