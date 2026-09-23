package contracttransactionbuilder

import (
	"dlt-ingress/src/main/dltingress/domain/common"
	"dlt-ingress/src/main/dltingress/port/contracttransactionbuilder"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDltIngressRegistry_Register(t *testing.T) {
	registry := contracttransactionbuilder.NewRegistry()
	expectedContractTransactionBuilder := new(ContractTransactionBuilderMock)
	dlt := common.EVM
	smartContractName := "testSmartContractName"
	registry.Register(dlt, smartContractName, expectedContractTransactionBuilder)

	contractTransactionBuilder, err := registry.GetContractTransactionBuilder(dlt, smartContractName)
	assert.Nil(t, err)
	assert.Equal(t, expectedContractTransactionBuilder, contractTransactionBuilder)
}

func TestDltIngressRegistry_GetNotFound(t *testing.T) {
	registry := contracttransactionbuilder.NewRegistry()
	expectedContractTransactionBuilder := new(ContractTransactionBuilderMock)
	invalidDlt := common.SVM
	smartContractName := "testSmartContractName"
	registry.Register(common.EVM, smartContractName, expectedContractTransactionBuilder)

	contractTransactionBuilder, err := registry.GetContractTransactionBuilder(invalidDlt, smartContractName)
	assert.Nil(t, contractTransactionBuilder)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("contract transaction builder not found for dlt %s and smart contract name %s", invalidDlt, smartContractName), err.Error())
}
