package transanctionsender

import (
	"dlt-ingress/src/main/dltingress/domain/common"
	"dlt-ingress/src/main/dltingress/port/transanctionsender"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDltIngressTransactionSenderRegistry_Register(t *testing.T) {
	registry := transanctionsender.NewRegistry()
	sender := new(TransactionSenderMock)
	dlt := common.EVM
	registry.Register(dlt, sender)

	result, err := registry.GetTransactionSender(dlt)
	assert.Nil(t, err)
	assert.Equal(t, sender, result)
}

func TestDltIngressTransactionSenderRegistry_GetNotFound(t *testing.T) {
	registry := transanctionsender.NewRegistry()
	registry.Register(common.EVM, new(TransactionSenderMock))

	result, err := registry.GetTransactionSender(common.SVM)
	assert.Nil(t, result)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("no transaction sender configured for dlt %s", common.SVM), err.Error())
}
