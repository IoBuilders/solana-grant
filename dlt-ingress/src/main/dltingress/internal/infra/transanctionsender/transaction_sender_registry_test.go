package transanctionsender

import (
	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDltIngressTransactionSenderRegistry_Register(t *testing.T) {
	registry := NewRegistry()
	sender := new(EvmTransactionSender)
	dlt := common.EVM
	registry.Register(dlt, sender)

	result, err := registry.GetTransactionSender(dlt)
	assert.Nil(t, err)
	assert.Equal(t, sender, result)
}

func TestDltIngressTransactionSenderRegistry_GetNotFound(t *testing.T) {
	registry := NewRegistry()
	registry.Register(common.EVM, new(EvmTransactionSender))

	result, err := registry.GetTransactionSender(common.SVM)
	assert.Nil(t, result)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("no transaction sender configured for dlt %s", common.SVM), err.Error())
}
