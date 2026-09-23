package contracttransactionbuilder

import (
	"dlt-ingress/src/main/dltingress/port/contracttransactionbuilder"
	"dlt-ingress/src/main/dltingress/port/portcommon"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
)

var abi = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"when\",\"type\":\"uint256\"}],\"name\":\"Withdrawal\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"simpleParam\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"testInt\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"testString\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"testAddress\",\"type\":\"address\"}],\"internalType\":\"struct Counter.TestBuildStruct\",\"name\":\"structParam\",\"type\":\"tuple\"}],\"name\":\"count\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"counter\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getCounter\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"
var chainId, _ = amount.NewFromString("1525")
var nonce, _ = amount.NewFromString("1")
var gasLimit, _ = amount.NewFromString("300000000")
var gasPrice, _ = amount.NewFromString("0")
var maxPriorityFeePerGas, _ = amount.NewFromString("10000")
var maxFeePerGas, _ = amount.NewFromString("50000")
var value, _ = amount.NewFromString("1")
var smartContractId = "0x92BDb699940f5a5Fb21C7D6EFA5a67C84B84A085"

func TestDltIngressEthGoContractTxBuilder_NewOk(t *testing.T) {
	contractTxBuilder, err := contracttransactionbuilder.NewEvmContractTransactionBuilder(abi)

	assert.Nil(t, err)
	assert.NotNil(t, contractTxBuilder)
}

func TestDltIngressEthGoContractTxBuilder_NewError(t *testing.T) {
	contractTxBuilder, err := contracttransactionbuilder.NewEvmContractTransactionBuilder("invalid abi")

	assert.Nil(t, contractTxBuilder)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "error parsing smart contract abi")
}

func TestDltIngressEthGoContractTxBuilder_MethodNotExistError(t *testing.T) {
	contractTxBuilder, _ := contracttransactionbuilder.NewEvmContractTransactionBuilder(abi)
	methodName := "grantRole"
	request := contracttransactionbuilder.NewEVMLegacyBuildTransactionRequest(
		"0x92BDb699940f5a5Fb21C7D6EFA5a67C84B84A085",
		smartContractId,
		methodName,
		map[string]any{},
		chainId,
		nonce,
		gasLimit,
		gasPrice,
		value,
	)
	resp, err := contractTxBuilder.BuildTransaction(request)
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("error building transaction: method %s does not exist", methodName), err.Error())
}

func TestDltIngressEthGoContractTxBuilder_DataError(t *testing.T) {
	contractTxBuilder, _ := contracttransactionbuilder.NewEvmContractTransactionBuilder(abi)
	request := contracttransactionbuilder.NewEVMLegacyBuildTransactionRequest(
		"0x92BDb699940f5a5Fb21C7D6EFA5a67C84B84A085",
		smartContractId,
		"count",
		map[string]any{},
		chainId,
		nonce,
		gasLimit,
		gasPrice,
		value,
	)
	resp, err := contractTxBuilder.BuildTransaction(request)
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "error encoding EVM transaction")
}

func TestDltIngressEthGoContractTxBuilder_LegacyOk(t *testing.T) {
	contractTxBuilder, _ := contracttransactionbuilder.NewEvmContractTransactionBuilder(abi)
	request := contracttransactionbuilder.NewEVMLegacyBuildTransactionRequest(
		"0x92BDb699940f5a5Fb21C7D6EFA5a67C84B84A085",
		smartContractId,
		"count",
		map[string]any{
			"simpleParam": 1,
			"structParam": map[string]any{
				"testString":  "test string",
				"testInt":     23456,
				"testAddress": "0x031f83263A8ddbbB90a80442aCC16E9eD5D50bD6",
			},
		},
		chainId,
		nonce,
		gasLimit,
		gasPrice,
		value,
	)
	resp, err := contractTxBuilder.BuildTransaction(request)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.EVMTransactionResponse)
	assert.Equal(t, portcommon.TransactionTypeLegacy, resp.EVMTransactionResponse.TransactionType)
	assert.Equal(t, chainId, resp.EVMTransactionResponse.ChainId)
	assert.Equal(t, nonce, resp.EVMTransactionResponse.Nonce)
	assert.Equal(t, gasLimit, resp.EVMTransactionResponse.GasLimit)
	assert.Equal(t, gasPrice, resp.EVMTransactionResponse.GasPrice)
	assert.Equal(t, "0xca221a58000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000005ba00000000000000000000000000000000000000000000000000000000000000060000000000000000000000000031f83263a8ddbbb90a80442acc16e9ed5d50bd6000000000000000000000000000000000000000000000000000000000000000b7465737420737472696e67000000000000000000000000000000000000000000", resp.EVMTransactionResponse.Data)
	assert.Equal(t, value, resp.EVMTransactionResponse.Value)
	assert.Nil(t, resp.EVMTransactionResponse.MaxPriorityFeePerGas)
	assert.Nil(t, resp.EVMTransactionResponse.MaxFeePerGas)
	assert.Equal(t, smartContractId, resp.EVMTransactionResponse.To)
}

func TestDltIngressEthGoContractTxBuilder_DynamicFeeOk(t *testing.T) {
	contractTxBuilder, _ := contracttransactionbuilder.NewEvmContractTransactionBuilder(abi)
	request := contracttransactionbuilder.NewEVMDynamicFeeBuildTransactionRequest(
		"0x92BDb699940f5a5Fb21C7D6EFA5a67C84B84A085",
		smartContractId,
		"count",
		map[string]any{
			"simpleParam": 1,
			"structParam": map[string]any{
				"testString":  "test string",
				"testInt":     23456,
				"testAddress": "0x031f83263A8ddbbB90a80442aCC16E9eD5D50bD6",
			},
		},
		chainId,
		nonce,
		gasLimit,
		maxPriorityFeePerGas,
		maxFeePerGas,
		value,
	)
	resp, err := contractTxBuilder.BuildTransaction(request)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.EVMTransactionResponse)
	assert.Equal(t, portcommon.TransactionTypeDynamicFee, resp.EVMTransactionResponse.TransactionType)
	assert.Equal(t, chainId, resp.EVMTransactionResponse.ChainId)
	assert.Equal(t, nonce, resp.EVMTransactionResponse.Nonce)
	assert.Equal(t, gasLimit, resp.EVMTransactionResponse.GasLimit)
	assert.Nil(t, resp.EVMTransactionResponse.GasPrice)
	assert.Equal(t, "0xca221a58000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000005ba00000000000000000000000000000000000000000000000000000000000000060000000000000000000000000031f83263a8ddbbb90a80442acc16e9ed5d50bd6000000000000000000000000000000000000000000000000000000000000000b7465737420737472696e67000000000000000000000000000000000000000000", resp.EVMTransactionResponse.Data)
	assert.Equal(t, value, resp.EVMTransactionResponse.Value)
	assert.Equal(t, maxPriorityFeePerGas, resp.EVMTransactionResponse.MaxPriorityFeePerGas)
	assert.Equal(t, maxFeePerGas, resp.EVMTransactionResponse.MaxFeePerGas)
	assert.Equal(t, smartContractId, resp.EVMTransactionResponse.To)
}

func TestDltIngressEthGoContractTxBuilder_OverrideGasLimit_SetsNewLimit(t *testing.T) {
	contractTxBuilder, _ := contracttransactionbuilder.NewEvmContractTransactionBuilder(abi)
	request := contracttransactionbuilder.NewEVMLegacyBuildTransactionRequest(
		"0x92BDb699940f5a5Fb21C7D6EFA5a67C84B84A085",
		smartContractId, "count",
		map[string]any{
			"simpleParam": 1,
			"structParam": map[string]any{"testString": "test string", "testInt": 23456, "testAddress": "0x031f83263A8ddbbB90a80442aCC16E9eD5D50bD6"},
		},
		chainId, nonce, gasLimit, gasPrice, value,
	)
	original, _ := contractTxBuilder.BuildTransaction(request)
	originalLimit := original.EVMTransactionResponse.GasLimit

	newLimit, _ := amount.NewFromString("999999")
	updated, err := contractTxBuilder.OverrideGasLimit(*request, *original, newLimit)

	assert.Nil(t, err)
	assert.Equal(t, newLimit, updated.EVMTransactionResponse.GasLimit)
	assert.Equal(t, originalLimit, original.EVMTransactionResponse.GasLimit, "original response must not be mutated")
}
