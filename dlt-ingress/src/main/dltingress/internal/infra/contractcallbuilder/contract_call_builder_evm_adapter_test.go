//go:build test

package contractcallbuilder

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ABI with a write method (count) and a view method (getCounter).
const callBuilderABI = `[
  {"inputs":[{"internalType":"uint256","name":"simpleParam","type":"uint256"},{"components":[{"internalType":"uint256","name":"testInt","type":"uint256"},{"internalType":"string","name":"testString","type":"string"},{"internalType":"address","name":"testAddress","type":"address"}],"internalType":"struct Counter.TestBuildStruct","name":"structParam","type":"tuple"}],"name":"count","outputs":[],"stateMutability":"nonpayable","type":"function"},
  {"inputs":[],"name":"getCounter","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},
  {"inputs":[],"name":"getInfo","outputs":[{"internalType":"uint256","name":"value","type":"uint256"},{"internalType":"string","name":"name","type":"string"}],"stateMutability":"view","type":"function"}
]`

const callBuilderContractId = "0x92BDb699940f5a5Fb21C7D6EFA5a67C84B84A085"

func newEvmCallBuilder(t *testing.T) *EvmContractCallBuilder {
	t.Helper()
	builder, err := NewEvmContractCallBuilder(callBuilderABI)
	require.NoError(t, err)
	return builder
}

func TestDltIngressEvmContractCallBuilder_New_Ok(t *testing.T) {
	builder, err := NewEvmContractCallBuilder(callBuilderABI)
	require.NoError(t, err)
	assert.NotNil(t, builder)
}

func TestDltIngressEvmContractCallBuilder_New_InvalidAbi(t *testing.T) {
	_, err := NewEvmContractCallBuilder("not json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing smart contract")
}

func TestDltIngressEvmContractCallBuilder_BuildCall_Ok(t *testing.T) {
	builder := newEvmCallBuilder(t)

	result, err := builder.BuildCall(&BuildCallRequest{
		SmartContractId: callBuilderContractId,
		MethodName:      "count",
		MethodArgs: map[string]any{
			"simpleParam": 1,
			"structParam": map[string]any{
				"testInt":     23456,
				"testString":  "hello",
				"testAddress": "0x031f83263A8ddbbB90a80442aCC16E9eD5D50bD6",
			},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, callBuilderContractId, result.To)
	assert.NotEmpty(t, result.Data)
	// Data must start with the 4-byte method selector (0xca221a58 for "count").
	assert.True(t, len(result.Data) > 10, "encoded call data too short")
}

func TestDltIngressEvmContractCallBuilder_BuildCall_MethodNotFound(t *testing.T) {
	builder := newEvmCallBuilder(t)

	result, err := builder.BuildCall(&BuildCallRequest{
		SmartContractId: callBuilderContractId,
		MethodName:      "nonexistent",
		MethodArgs:      map[string]any{},
	})

	assert.Nil(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestDltIngressEvmContractCallBuilder_BuildCall_EncodingError(t *testing.T) {
	builder := newEvmCallBuilder(t)

	// count requires simpleParam and structParam; passing nothing triggers an encode error.
	result, err := builder.BuildCall(&BuildCallRequest{
		SmartContractId: callBuilderContractId,
		MethodName:      "count",
		MethodArgs:      map[string]any{},
	})

	assert.Nil(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error encoding EVM call")
}

func TestDltIngressEvmContractCallBuilder_DecodeResult_Ok(t *testing.T) {
	builder := newEvmCallBuilder(t)

	// ABI-encode uint256(42) as a hex string (as returned by eth_call).
	// Unnamed outputs are keyed by their positional index as a string ("0").
	decoded, err := builder.DecodeResult("getCounter", "0x000000000000000000000000000000000000000000000000000000000000002a")

	require.NoError(t, err)
	require.NotNil(t, decoded)
	assert.Equal(t, 1, len(decoded))
	assert.Equal(t, big.NewInt(42), decoded["0"])
}

func TestDltIngressEvmContractCallBuilder_DecodeResult_MultipleNamedOutputs(t *testing.T) {
	builder := newEvmCallBuilder(t)

	// ABI-encode (uint256(7), string("Alice")) — two named outputs → decoded as map[string]any.
	decoded, err := builder.DecodeResult("getInfo",
		"0x"+
			"0000000000000000000000000000000000000000000000000000000000000007"+ // value = 7
			"0000000000000000000000000000000000000000000000000000000000000040"+ // offset to string (64 bytes)
			"0000000000000000000000000000000000000000000000000000000000000005"+ // string length = 5
			"416c696365000000000000000000000000000000000000000000000000000000") // "Alice" padded

	require.NoError(t, err)
	require.NotNil(t, decoded)
	assert.Equal(t, 2, len(decoded))
	assert.Equal(t, big.NewInt(7), decoded["value"])
	assert.Equal(t, "Alice", decoded["name"])
}

func TestDltIngressEvmContractCallBuilder_DecodeResult_InvalidHex(t *testing.T) {
	builder := newEvmCallBuilder(t)

	decoded, err := builder.DecodeResult("getCounter", "not-hex")

	assert.Nil(t, decoded)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error decoding call result hex")
}

func TestDltIngressEvmContractCallBuilder_DecodeResult_MethodNotFound(t *testing.T) {
	builder := newEvmCallBuilder(t)

	decoded, err := builder.DecodeResult("nonexistent", "0x00")

	assert.Nil(t, decoded)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestDltIngressEvmContractCallBuilder_ImplementsPortInterface(t *testing.T) {
	var _ Port = (*EvmContractCallBuilder)(nil)
}
