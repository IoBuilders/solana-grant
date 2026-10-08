package contracttransactionbuilder

import (
	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDltIngressRegistry_Register(t *testing.T) {
	registry := NewRegistry()
	expectedContractTransactionBuilder := new(EvmContractTransactionBuilder)
	dlt := common.EVM
	smartContractName := "testSmartContractName"
	registry.Register(dlt, smartContractName, expectedContractTransactionBuilder)

	contractTransactionBuilder, err := registry.GetContractTransactionBuilder(dlt, smartContractName)
	assert.Nil(t, err)
	assert.Equal(t, expectedContractTransactionBuilder, contractTransactionBuilder)
}

func TestDltIngressRegistry_GetNotFound(t *testing.T) {
	registry := NewRegistry()
	expectedContractTransactionBuilder := new(EvmContractTransactionBuilder)
	invalidDlt := common.SVM
	smartContractName := "testSmartContractName"
	registry.Register(common.EVM, smartContractName, expectedContractTransactionBuilder)

	contractTransactionBuilder, err := registry.GetContractTransactionBuilder(invalidDlt, smartContractName)
	assert.Nil(t, contractTransactionBuilder)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("contract transaction builder not found for dlt %s and smart contract name %s", invalidDlt, smartContractName), err.Error())
}

func TestDltIngressRegistry_GetNativeTransferContractTransactionBuilder(t *testing.T) {
	registry := NewRegistry()
	nativeTransferContractTransactionBuilder := registry.GetNativeTransferTransactionBuilder()
	assert.NotNil(t, nativeTransferContractTransactionBuilder)
	assert.IsType(t, &NativeTransferContractTransactionBuilder{}, nativeTransferContractTransactionBuilder)
}
