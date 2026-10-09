package contracttransactionbuilder

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var batchAbi = `[
	{"inputs":[{"internalType":"uint256","name":"simpleParam","type":"uint256"}],"name":"singleArg","outputs":[],"stateMutability":"nonpayable","type":"function"},
	{"inputs":[],"name":"noArgs","outputs":[],"stateMutability":"nonpayable","type":"function"},
	{"inputs":[{"internalType":"address","name":"_target","type":"address"},{"internalType":"bytes[]","name":"_callData","type":"bytes[]"}],"name":"batch","outputs":[],"stateMutability":"nonpayable","type":"function"}
]`

func TestDltIngressEvmEncodeCallData_ReturnsSelectorAndArgs(t *testing.T) {
	builder, err := NewEvmContractTransactionBuilder(batchAbi)
	require.NoError(t, err)

	encoded, err := builder.EncodeCallData("singleArg", map[string]any{"simpleParam": 7})

	require.NoError(t, err)
	// 4 bytes of selector plus one 32-byte word for the uint256.
	assert.Len(t, encoded, 36)
	assert.Equal(t,
		"0000000000000000000000000000000000000000000000000000000000000007",
		hex.EncodeToString(encoded[4:]),
	)
}

func TestDltIngressEvmEncodeCallData_MethodWithoutArgsIsSelectorOnly(t *testing.T) {
	builder, err := NewEvmContractTransactionBuilder(batchAbi)
	require.NoError(t, err)

	encoded, err := builder.EncodeCallData("noArgs", map[string]any{})

	require.NoError(t, err)
	assert.Len(t, encoded, 4)
}

func TestDltIngressEvmEncodeCallData_SelectorDiffersPerMethod(t *testing.T) {
	builder, err := NewEvmContractTransactionBuilder(batchAbi)
	require.NoError(t, err)

	single, err := builder.EncodeCallData("singleArg", map[string]any{"simpleParam": 1})
	require.NoError(t, err)
	none, err := builder.EncodeCallData("noArgs", map[string]any{})
	require.NoError(t, err)

	assert.NotEqual(t, single[:4], none[:4])
}

func TestDltIngressEvmEncodeCallData_MethodNotExistError(t *testing.T) {
	builder, err := NewEvmContractTransactionBuilder(batchAbi)
	require.NoError(t, err)

	encoded, err := builder.EncodeCallData("doesNotExist", map[string]any{})

	assert.Nil(t, encoded)
	require.Error(t, err)
	assert.Equal(t, "error encoding EVM call data: method doesNotExist does not exist", err.Error())
}

func TestDltIngressEvmEncodeCallData_WrongArgsError(t *testing.T) {
	builder, err := NewEvmContractTransactionBuilder(batchAbi)
	require.NoError(t, err)

	encoded, err := builder.EncodeCallData("singleArg", map[string]any{"unknownParam": 1})

	assert.Nil(t, encoded)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error encoding EVM call data for method singleArg")
}

func TestDltIngressEvmEncodeCallData_ResultFitsIntoBytesArrayArgument(t *testing.T) {
	builder, err := NewEvmContractTransactionBuilder(batchAbi)
	require.NoError(t, err)

	first, err := builder.EncodeCallData("singleArg", map[string]any{"simpleParam": 1})
	require.NoError(t, err)
	second, err := builder.EncodeCallData("noArgs", map[string]any{})
	require.NoError(t, err)

	outer, err := builder.EncodeCallData("batch", map[string]any{
		"_target":   "0x92BDb699940f5a5Fb21C7D6EFA5a67C84B84A085",
		"_callData": [][]byte{first, second},
	})

	require.NoError(t, err)
	assert.Greater(t, len(outer), 4)
}
