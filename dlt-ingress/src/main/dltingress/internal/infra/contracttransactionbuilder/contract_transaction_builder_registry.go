package contracttransactionbuilder

import (
	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"fmt"
)

type Registry struct {
	contractTransactionBuilders              map[common.Dlt]map[string]Port
	nativeTransferContractTransactionBuilder Port
}

func NewRegistry() *Registry {
	return &Registry{
		contractTransactionBuilders:              make(map[common.Dlt]map[string]Port),
		nativeTransferContractTransactionBuilder: &NativeTransferContractTransactionBuilder{},
	}
}

func (r *Registry) Register(dlt common.Dlt, smartContractName string, contractTransactionBuilder Port) {
	if r.contractTransactionBuilders[dlt] == nil {
		r.contractTransactionBuilders[dlt] = make(map[string]Port)
	}
	r.contractTransactionBuilders[dlt][smartContractName] = contractTransactionBuilder
}

func (r *Registry) GetContractTransactionBuilder(dlt common.Dlt, smartContractName string) (Port, error) {
	if contractTransactionBuilder, found := r.contractTransactionBuilders[dlt][smartContractName]; found {
		return contractTransactionBuilder, nil
	}
	return nil, fmt.Errorf("contract transaction builder not found for dlt %s and smart contract name %s", dlt, smartContractName)
}

func (r *Registry) GetNativeTransferTransactionBuilder() Port {
	return r.nativeTransferContractTransactionBuilder
}
