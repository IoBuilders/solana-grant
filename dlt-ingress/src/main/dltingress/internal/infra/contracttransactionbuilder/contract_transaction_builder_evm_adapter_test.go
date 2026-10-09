package contracttransactionbuilder

import (
	"errors"
	"fmt"
	"testing"

	"dlt-ingress/src/main/dltingress/internal/domain/domainerrors"
	"dlt-ingress/src/main/dltingress/internal/infra/portcommon"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
)

var exampleAbi = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"when\",\"type\":\"uint256\"}],\"name\":\"Withdrawal\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"simpleParam\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"testInt\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"testString\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"testAddress\",\"type\":\"address\"}],\"internalType\":\"struct Counter.TestBuildStruct\",\"name\":\"structParam\",\"type\":\"tuple\"}],\"name\":\"count\",\"outputs\":[],\"stateMutexampleAbility\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"counter\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutexampleAbility\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getCounter\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutexampleAbility\":\"view\",\"type\":\"function\"}]"
var chainId, _ = amount.NewFromString("1525")
var nonce, _ = amount.NewFromString("1")
var gasLimit, _ = amount.NewFromString("300000000")
var gasPrice, _ = amount.NewFromString("0")
var maxPriorityFeePerGas, _ = amount.NewFromString("10000")
var maxFeePerGas, _ = amount.NewFromString("50000")
var value, _ = amount.NewFromString("1")
var smartContractId = "0x92BDb699940f5a5Fb21C7D6EFA5a67C84B84A085"

func TestDltIngressEthGoContractTxBuilder_NewOk(t *testing.T) {
	contractTxBuilder, err := NewEvmContractTransactionBuilder(exampleAbi)

	assert.Nil(t, err)
	assert.NotNil(t, contractTxBuilder)
}

func TestDltIngressEthGoContractTxBuilder_NewError(t *testing.T) {
	contractTxBuilder, err := NewEvmContractTransactionBuilder("invalid abi")

	assert.Nil(t, contractTxBuilder)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "error parsing smart contract abi")
}

func TestDltIngressEthGoContractTxBuilder_MethodNotExistError(t *testing.T) {
	contractTxBuilder, _ := NewEvmContractTransactionBuilder(exampleAbi)
	methodName := "grantRole"
	request := NewEVMLegacyBuildTransactionRequest(
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
	assert.Equal(t, fmt.Sprintf("error encoding EVM call data: method %s does not exist", methodName), err.Error())
}

func TestDltIngressEthGoContractTxBuilder_DataError(t *testing.T) {
	contractTxBuilder, _ := NewEvmContractTransactionBuilder(exampleAbi)
	request := NewEVMLegacyBuildTransactionRequest(
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
	assert.Contains(t, err.Error(), "error encoding EVM call data for method count")
}

func TestDltIngressEthGoContractTxBuilder_LegacyOk(t *testing.T) {
	contractTxBuilder, _ := NewEvmContractTransactionBuilder(exampleAbi)
	request := NewEVMLegacyBuildTransactionRequest(
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
	contractTxBuilder, _ := NewEvmContractTransactionBuilder(exampleAbi)
	request := NewEVMDynamicFeeBuildTransactionRequest(
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
	contractTxBuilder, _ := NewEvmContractTransactionBuilder(exampleAbi)
	request := NewEVMLegacyBuildTransactionRequest(
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

func TestDltIngressEthGoContractTxBuilder_OverrideCuPrice_NotSupported(t *testing.T) {
	contractTxBuilder, _ := NewEvmContractTransactionBuilder(exampleAbi)
	cuPrice, _ := amount.NewFromString("5000")

	updated, err := contractTxBuilder.OverrideCuPrice(BuildTransactionRequest{}, portcommon.TransactionResponse{}, cuPrice)

	assert.Nil(t, updated)
	assert.ErrorContains(t, err, "not supported on EVM")
}

func TestDltIngressEthGoContractTxBuilder_RejectsASmartContractIdThatIsNotAnEvmAddress(t *testing.T) {
	contractTxBuilder, _ := NewEvmContractTransactionBuilder(exampleAbi)

	for name, invalidId := range map[string]string{
		"empty":             "",
		"not hex":           "not-an-address",
		"too short":         "0x123",
		"too long":          smartContractId + "00",
		"non hex digit":     "0x92BDb699940f5a5Fb21C7D6EFA5a67C84B84A08Z",
		"solana program id": "11111111111111111111111111111112",
	} {
		t.Run(name, func(t *testing.T) {
			request := NewEVMLegacyBuildTransactionRequest(
				"0x92BDb699940f5a5Fb21C7D6EFA5a67C84B84A085",
				invalidId,
				"getCounter",
				map[string]any{},
				chainId,
				nonce,
				gasLimit,
				gasPrice,
				value,
			)

			resp, err := contractTxBuilder.BuildTransaction(request)

			assert.Nil(t, resp)
			var domainErr coreerror.DomainError
			require.True(t, errors.As(err, &domainErr), "expected a domain error, got %v", err)
			assert.Equal(t, domainerrors.ErrorCodeInvalidSmartContractAddress, domainErr.ErrorCode())
		})
	}
}

func TestDltIngressEthGoContractTxBuilder_RejectsAnInvalidSmartContractIdOnDynamicFeeToo(t *testing.T) {
	contractTxBuilder, _ := NewEvmContractTransactionBuilder(exampleAbi)
	request := NewEVMDynamicFeeBuildTransactionRequest(
		"0x92BDb699940f5a5Fb21C7D6EFA5a67C84B84A085",
		"",
		"getCounter",
		map[string]any{},
		chainId,
		nonce,
		gasLimit,
		maxPriorityFeePerGas,
		maxFeePerGas,
		value,
	)

	_, err := contractTxBuilder.BuildTransaction(request)

	var domainErr coreerror.DomainError
	require.True(t, errors.As(err, &domainErr), "expected a domain error, got %v", err)
	assert.Equal(t, domainerrors.ErrorCodeInvalidSmartContractAddress, domainErr.ErrorCode())
}
