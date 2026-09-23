package transactiongasestimator

import (
	"dlt-ingress/src/main/dltingress/domain/common"
	"dlt-ingress/src/main/dltingress/port/transactiongasestimator"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDltIngressTransactionGasEstimatorRegistry_Register(t *testing.T) {
	registry := transactiongasestimator.NewRegistry()
	sender := new(TransactionGasEstimatorMock)
	dlt := common.EVM
	registry.Register(dlt, sender)

	result, err := registry.GetTransactionGasEstimator(dlt)
	assert.Nil(t, err)
	assert.Equal(t, sender, result)
}

func TestDltIngressTransactionGasEstimatorRegistry_GetNotFound(t *testing.T) {
	registry := transactiongasestimator.NewRegistry()
	registry.Register(common.EVM, new(TransactionGasEstimatorMock))

	result, err := registry.GetTransactionGasEstimator(common.SVM)
	assert.Nil(t, result)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("no transaction estimator configured for dlt %s", common.SVM), err.Error())
}
