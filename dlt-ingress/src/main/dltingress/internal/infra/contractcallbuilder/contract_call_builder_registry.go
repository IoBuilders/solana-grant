package contractcallbuilder

import (
	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"fmt"
)

type Registry struct {
	contractCallBuilders map[common.Dlt]map[string]Port
}

func NewRegistry() *Registry {
	return &Registry{
		contractCallBuilders: make(map[common.Dlt]map[string]Port),
	}
}

func (r *Registry) Register(dlt common.Dlt, smartContractName string, contractCallBuilder Port) {
	if r.contractCallBuilders[dlt] == nil {
		r.contractCallBuilders[dlt] = make(map[string]Port)
	}
	r.contractCallBuilders[dlt][smartContractName] = contractCallBuilder
}

func (r *Registry) GetContractCallBuilder(dlt common.Dlt, smartContractName string) (Port, error) {
	if builder, found := r.contractCallBuilders[dlt][smartContractName]; found {
		return builder, nil
	}
	return nil, fmt.Errorf("contract call builder not found for dlt %s and smart contract name %s", dlt, smartContractName)
}
