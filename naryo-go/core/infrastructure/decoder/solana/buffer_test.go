//go:build test

package solanadecoder

import (
	"math/big"
	"testing"

	"github.com/mr-tron/base58"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter/solana"
)

func mustUint(t *testing.T, bitSize int) solana.ParameterDefinition {
	t.Helper()
	def, err := solana.NewUintParameterDefinition(0, bitSize)
	require.NoError(t, err)
	return def
}

func mustInt(t *testing.T, bitSize int) solana.ParameterDefinition {
	t.Helper()
	def, err := solana.NewIntParameterDefinition(0, bitSize)
	require.NoError(t, err)
	return def
}

func TestDecodeBuffer_Scalars(t *testing.T) {
	t.Run("Uint widths", func(t *testing.T) {
		cases := []struct {
			bitSize int
			data    []byte
			want    *big.Int
		}{
			{8, []byte{42}, big.NewInt(42)},
			{16, []byte{0x2C, 0x01}, big.NewInt(300)},
			{32, []byte{0x00, 0x00, 0x00, 0x01}, new(big.Int).Lsh(big.NewInt(1), 24)},
		}

		for _, c := range cases {
			def, err := solana.NewUintParameterDefinition(0, c.bitSize)
			require.NoError(t, err)
			params, err := DecodeBuffer(c.data, []solana.ParameterDefinition{def})
			require.NoError(t, err)
			require.Len(t, params, 1)
			assert.Equal(t, parameter.TypeUint, params[0].Type())
			assert.Equal(t, c.want, params[0].Value())
		}
	})

	t.Run("Uint128 crosses the 64-bit boundary", func(t *testing.T) {
		data := make([]byte, 16)
		data[8] = 1 // 2^64
		params, err := DecodeBuffer(data, []solana.ParameterDefinition{mustUint(t, 128)})
		require.NoError(t, err)
		expected := new(big.Int).Lsh(big.NewInt(1), 64)
		assert.Equal(t, expected, params[0].Value())
	})

	t.Run("Uint128 max value", func(t *testing.T) {
		data := make([]byte, 16)
		for i := range data {
			data[i] = 0xFF
		}
		params, err := DecodeBuffer(data, []solana.ParameterDefinition{mustUint(t, 128)})
		require.NoError(t, err)
		expected := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))
		assert.Equal(t, expected, params[0].Value())
	})

	t.Run("Int64 positive", func(t *testing.T) {
		data := []byte{0x05, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
		params, err := DecodeBuffer(data, []solana.ParameterDefinition{mustInt(t, 64)})
		require.NoError(t, err)
		assert.Equal(t, parameter.TypeInt, params[0].Type())
		assert.Equal(t, big.NewInt(5), params[0].Value())
	})

	t.Run("Int64 negative", func(t *testing.T) {
		data := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF} // -1
		params, err := DecodeBuffer(data, []solana.ParameterDefinition{mustInt(t, 64)})
		require.NoError(t, err)
		assert.Equal(t, big.NewInt(-1), params[0].Value())
	})

	t.Run("Int8 negative", func(t *testing.T) {
		data := []byte{0xFF} // -1
		params, err := DecodeBuffer(data, []solana.ParameterDefinition{mustInt(t, 8)})
		require.NoError(t, err)
		assert.Equal(t, big.NewInt(-1), params[0].Value())
	})

	t.Run("Bool", func(t *testing.T) {
		def, err := solana.NewBoolParameterDefinition(0)
		require.NoError(t, err)
		params, err := DecodeBuffer([]byte{1}, []solana.ParameterDefinition{def})
		require.NoError(t, err)
		assert.Equal(t, parameter.TypeBool, params[0].Type())
		assert.Equal(t, true, params[0].Value())
	})

	t.Run("Float32", func(t *testing.T) {
		def, err := solana.NewFloatParameterDefinition(0, 32)
		require.NoError(t, err)
		// 1.5 as IEEE-754 float32 little-endian
		data := []byte{0x00, 0x00, 0xC0, 0x3F}
		params, err := DecodeBuffer(data, []solana.ParameterDefinition{def})
		require.NoError(t, err)
		assert.Equal(t, parameter.TypeFloat, params[0].Type())
		assert.InDelta(t, 1.5, params[0].Value(), 0.0001)
	})

	t.Run("Float64", func(t *testing.T) {
		def, err := solana.NewFloatParameterDefinition(0, 64)
		require.NoError(t, err)
		// 2.5 as IEEE-754 float64 little-endian
		data := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0x40}
		params, err := DecodeBuffer(data, []solana.ParameterDefinition{def})
		require.NoError(t, err)
		assert.InDelta(t, 2.5, params[0].Value(), 0.0001)
	})

	t.Run("PublicKey", func(t *testing.T) {
		def, err := solana.NewPublicKeyParameterDefinition(0)
		require.NoError(t, err)
		raw := make([]byte, 32)
		for i := range raw {
			raw[i] = byte(i)
		}
		params, err := DecodeBuffer(raw, []solana.ParameterDefinition{def})
		require.NoError(t, err)
		assert.Equal(t, parameter.TypePublicKey, params[0].Type())
		encoded, ok := params[0].Value().(string)
		require.True(t, ok)
		decoded, err := base58.Decode(encoded)
		require.NoError(t, err)
		assert.Equal(t, raw, decoded)
	})

	t.Run("BytesFixed", func(t *testing.T) {
		def, err := solana.NewBytesFixedParameterDefinition(0, 4)
		require.NoError(t, err)
		data := []byte{1, 2, 3, 4}
		params, err := DecodeBuffer(data, []solana.ParameterDefinition{def})
		require.NoError(t, err)
		assert.Equal(t, parameter.TypeBytesFixed, params[0].Type())
		assert.Equal(t, data, params[0].Value())
	})

	t.Run("String", func(t *testing.T) {
		def, err := solana.NewStringParameterDefinition(0)
		require.NoError(t, err)
		// u32 LE length prefix (5) + "hello"
		data := append([]byte{5, 0, 0, 0}, []byte("hello")...)
		params, err := DecodeBuffer(data, []solana.ParameterDefinition{def})
		require.NoError(t, err)
		assert.Equal(t, parameter.TypeString, params[0].Type())
		assert.Equal(t, "hello", params[0].Value())
	})

	t.Run("EmptyString", func(t *testing.T) {
		def, err := solana.NewStringParameterDefinition(0)
		require.NoError(t, err)
		data := []byte{0, 0, 0, 0}
		params, err := DecodeBuffer(data, []solana.ParameterDefinition{def})
		require.NoError(t, err)
		assert.Equal(t, "", params[0].Value())
	})

	t.Run("Bytes", func(t *testing.T) {
		def, err := solana.NewBytesParameterDefinition(0)
		require.NoError(t, err)
		data := append([]byte{3, 0, 0, 0}, []byte{9, 8, 7}...)
		params, err := DecodeBuffer(data, []solana.ParameterDefinition{def})
		require.NoError(t, err)
		assert.Equal(t, parameter.TypeBytes, params[0].Type())
		assert.Equal(t, []byte{9, 8, 7}, params[0].Value())
	})
}

func TestDecodeBuffer_SequentialMultiField(t *testing.T) {
	data := []byte{
		7,                                              // U8
		1,                                              // BOOL true
		0x0A, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // U64 = 10
	}
	uint8Def, err := solana.NewUintParameterDefinition(0, 8)
	require.NoError(t, err)
	boolDef, err := solana.NewBoolParameterDefinition(1)
	require.NoError(t, err)
	uint64Def, err := solana.NewUintParameterDefinition(2, 64)
	require.NoError(t, err)

	params, err := DecodeBuffer(data, []solana.ParameterDefinition{uint8Def, boolDef, uint64Def})
	require.NoError(t, err)
	require.Len(t, params, 3)
	assert.Equal(t, big.NewInt(7), params[0].Value())
	assert.Equal(t, true, params[1].Value())
	assert.Equal(t, big.NewInt(10), params[2].Value())
}

func TestDecodeBuffer_Array(t *testing.T) {
	elem, err := solana.NewUintParameterDefinition(0, 8)
	require.NoError(t, err)

	t.Run("FixedArray", func(t *testing.T) {
		length := 3
		def, err := solana.NewArrayParameterDefinition(0, elem, &length)
		require.NoError(t, err)
		data := []byte{1, 2, 3}
		params, err := DecodeBuffer(data, []solana.ParameterDefinition{def})
		require.NoError(t, err)
		require.Len(t, params, 1)
		assert.Equal(t, parameter.TypeArray, params[0].Type())
		elements, ok := params[0].Value().([]parameter.ContractEventParameter)
		require.True(t, ok)
		require.Len(t, elements, 3)
		assert.Equal(t, big.NewInt(1), elements[0].Value())
		assert.Equal(t, big.NewInt(2), elements[1].Value())
		assert.Equal(t, big.NewInt(3), elements[2].Value())
	})

	t.Run("DynamicVec", func(t *testing.T) {
		def, err := solana.NewArrayParameterDefinition(0, elem, nil)
		require.NoError(t, err)
		// u32 LE count (2) + 2 elements
		data := []byte{2, 0, 0, 0, 9, 8}
		params, err := DecodeBuffer(data, []solana.ParameterDefinition{def})
		require.NoError(t, err)
		elements, ok := params[0].Value().([]parameter.ContractEventParameter)
		require.True(t, ok)
		require.Len(t, elements, 2)
		assert.Equal(t, big.NewInt(9), elements[0].Value())
		assert.Equal(t, big.NewInt(8), elements[1].Value())
	})

	t.Run("EmptyVec", func(t *testing.T) {
		def, err := solana.NewArrayParameterDefinition(0, elem, nil)
		require.NoError(t, err)
		// u32 LE count (0), no elements follow — a legitimate empty Vec<T>
		data := []byte{0, 0, 0, 0}
		params, err := DecodeBuffer(data, []solana.ParameterDefinition{def})
		require.NoError(t, err)
		require.Len(t, params, 1)
		elements, ok := params[0].Value().([]parameter.ContractEventParameter)
		require.True(t, ok)
		assert.Empty(t, elements)
	})

	t.Run("VecCountExceedsBuffer", func(t *testing.T) {
		def, err := solana.NewArrayParameterDefinition(0, elem, nil)
		require.NoError(t, err)
		// declares 1000 elements but the buffer has none
		data := []byte{0xE8, 0x03, 0x00, 0x00}
		_, err = DecodeBuffer(data, []solana.ParameterDefinition{def})
		require.Error(t, err)
		assert.ErrorIs(t, err, domainerrors.ErrDecoding)
	})
}

func TestDecodeBuffer_Struct(t *testing.T) {
	field0, err := solana.NewUintParameterDefinition(0, 8)
	require.NoError(t, err)
	field1, err := solana.NewBoolParameterDefinition(1)
	require.NoError(t, err)
	structDef, err := solana.NewStructParameterDefinition(0, []solana.ParameterDefinition{field0, field1})
	require.NoError(t, err)

	data := []byte{42, 1}
	params, err := DecodeBuffer(data, []solana.ParameterDefinition{structDef})
	require.NoError(t, err)
	require.Len(t, params, 1)
	assert.Equal(t, parameter.TypeStruct, params[0].Type())
	fields, ok := params[0].Value().([]parameter.ContractEventParameter)
	require.True(t, ok)
	require.Len(t, fields, 2)
	assert.Equal(t, big.NewInt(42), fields[0].Value())
	assert.Equal(t, true, fields[1].Value())
}

func TestDecodeBuffer_Option(t *testing.T) {
	// Inner and the Option wrapping it deliberately have different
	// positions, to prove the decoded inner value reports its own
	// declared position rather than inheriting the outer Option's.
	inner, err := solana.NewUintParameterDefinition(3, 8)
	require.NoError(t, err)
	optDef, err := solana.NewOptionParameterDefinition(5, inner)
	require.NoError(t, err)

	t.Run("None", func(t *testing.T) {
		params, err := DecodeBuffer([]byte{0}, []solana.ParameterDefinition{optDef})
		require.NoError(t, err)
		assert.Equal(t, parameter.TypeOption, params[0].Type())
		assert.Equal(t, 5, params[0].Position())
		assert.Nil(t, params[0].Value())
	})

	t.Run("Some", func(t *testing.T) {
		params, err := DecodeBuffer([]byte{1, 42}, []solana.ParameterDefinition{optDef})
		require.NoError(t, err)
		assert.Equal(t, 5, params[0].Position())
		innerParam, ok := params[0].Value().(parameter.ContractEventParameter)
		require.True(t, ok)
		assert.Equal(t, big.NewInt(42), innerParam.Value())
		assert.Equal(t, 3, innerParam.Position())
	})

	t.Run("InvalidTag", func(t *testing.T) {
		_, err := DecodeBuffer([]byte{2, 42}, []solana.ParameterDefinition{optDef})
		require.Error(t, err)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}

func TestDecodeBuffer_Underflow(t *testing.T) {
	cases := []struct {
		name string
		def  func(t *testing.T) solana.ParameterDefinition
		data []byte
	}{
		{"Uint8", func(t *testing.T) solana.ParameterDefinition { return mustUint(t, 8) }, nil},
		{"Uint64", func(t *testing.T) solana.ParameterDefinition { return mustUint(t, 64) }, make([]byte, 7)},
		{"Uint128", func(t *testing.T) solana.ParameterDefinition { return mustUint(t, 128) }, make([]byte, 15)},
		{"Int64", func(t *testing.T) solana.ParameterDefinition { return mustInt(t, 64) }, make([]byte, 7)},
		{"Bool", func(t *testing.T) solana.ParameterDefinition {
			d, err := solana.NewBoolParameterDefinition(0)
			require.NoError(t, err)
			return d
		}, nil},
		{"PublicKey", func(t *testing.T) solana.ParameterDefinition {
			d, err := solana.NewPublicKeyParameterDefinition(0)
			require.NoError(t, err)
			return d
		}, make([]byte, 31)},
		{"BytesFixed", func(t *testing.T) solana.ParameterDefinition {
			d, err := solana.NewBytesFixedParameterDefinition(0, 4)
			require.NoError(t, err)
			return d
		}, make([]byte, 3)},
		{"StringMissingPrefix", func(t *testing.T) solana.ParameterDefinition {
			d, err := solana.NewStringParameterDefinition(0)
			require.NoError(t, err)
			return d
		}, make([]byte, 3)},
		{"StringLengthExceedsBuffer", func(t *testing.T) solana.ParameterDefinition {
			d, err := solana.NewStringParameterDefinition(0)
			require.NoError(t, err)
			return d
		}, []byte{100, 0, 0, 0}}, // declares 100 bytes, none present
		{"BytesLengthExceedsBuffer", func(t *testing.T) solana.ParameterDefinition {
			d, err := solana.NewBytesParameterDefinition(0)
			require.NoError(t, err)
			return d
		}, []byte{100, 0, 0, 0}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			params, err := DecodeBuffer(c.data, []solana.ParameterDefinition{c.def(t)})
			require.Error(t, err)
			assert.Nil(t, params)
			assert.ErrorIs(t, err, domainerrors.ErrDecoding)

			var domainErr domainerrors.DomainError
			require.ErrorAs(t, err, &domainErr)
			assert.Equal(t, domainerrors.ErrorCodeBufferUnderflow, domainErr.ErrorCode())
		})
	}

	t.Run("TruncatedStructField", func(t *testing.T) {
		field0, err := solana.NewUintParameterDefinition(0, 64)
		require.NoError(t, err)
		structDef, err := solana.NewStructParameterDefinition(0, []solana.ParameterDefinition{field0})
		require.NoError(t, err)
		_, err = DecodeBuffer(make([]byte, 3), []solana.ParameterDefinition{structDef})
		require.Error(t, err)
		assert.ErrorIs(t, err, domainerrors.ErrDecoding)
	})

	t.Run("TruncatedArrayElement", func(t *testing.T) {
		elem, err := solana.NewUintParameterDefinition(0, 64)
		require.NoError(t, err)
		length := 2
		// Position 3 (distinct from the element loop index) so the error
		// message can be checked to report the array's own position, not
		// just the index of the element that ran out of bytes.
		arrDef, err := solana.NewArrayParameterDefinition(3, elem, &length)
		require.NoError(t, err)
		_, err = DecodeBuffer(make([]byte, 8), []solana.ParameterDefinition{arrDef}) // only 1 element's worth
		require.Error(t, err)
		assert.ErrorIs(t, err, domainerrors.ErrDecoding)
		assert.ErrorContains(t, err, "array at position 3")
		assert.ErrorContains(t, err, "element 1")
	})
}

func TestDecodeBuffer_EmptyDefs(t *testing.T) {
	params, err := DecodeBuffer([]byte{1, 2, 3}, nil)
	require.NoError(t, err)
	assert.Empty(t, params)
}

func TestDecodeBuffer_UnsupportedDefinition(t *testing.T) {
	_, err := DecodeBuffer([]byte{1}, []solana.ParameterDefinition{unsupportedDefinition{}})
	require.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrValidation)
}

// unsupportedDefinition is a hand-written ParameterDefinition test double
// that isn't one of solana's concrete types, to exercise DecodeBuffer's
// default case.
type unsupportedDefinition struct{}

func (unsupportedDefinition) Type() parameter.Type { return parameter.TypeBool }
func (unsupportedDefinition) Position() int        { return 0 }
