package transactiongasestimator

import (
	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"fmt"
)

type Registry struct {
	estimators map[common.Dlt]Port
}

func NewRegistry() *Registry {
	return &Registry{
		estimators: make(map[common.Dlt]Port),
	}
}

func (r *Registry) Register(dlt common.Dlt, estimator Port) {
	r.estimators[dlt] = estimator
}

func (r *Registry) GetTransactionGasEstimator(dlt common.Dlt) (Port, error) {
	estimator, found := r.estimators[dlt]
	if !found {
		return nil, fmt.Errorf("no transaction estimator configured for dlt %s", dlt)
	}
	return estimator, nil
}
