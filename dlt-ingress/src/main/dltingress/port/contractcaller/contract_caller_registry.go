package contractcaller

import (
	"dlt-ingress/src/main/dltingress/domain/common"
	"fmt"
)

type Registry struct {
	callers map[common.Dlt]Port
}

func NewRegistry() *Registry {
	return &Registry{callers: make(map[common.Dlt]Port)}
}

func (r *Registry) Register(dlt common.Dlt, caller Port) {
	r.callers[dlt] = caller
}

func (r *Registry) GetContractCaller(dlt common.Dlt) (Port, error) {
	caller, found := r.callers[dlt]
	if !found {
		return nil, fmt.Errorf("no contract caller configured for dlt %s", dlt)
	}
	return caller, nil
}
