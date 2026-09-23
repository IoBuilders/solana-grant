//go:build test

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter/solana"
)

func TestParseAnchorSignature(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		eventName, params, err := parseAnchorSignature("Transfer(uint16,string,bytes)")

		require.NoError(t, err)
		assert.Equal(t, "Transfer", eventName)
		require.Len(t, params, 3)
		uintDef, ok := params[0].(solana.UintParameterDefinition)
		require.True(t, ok)
		assert.Equal(t, 16, uintDef.BitSize)
		assert.IsType(t, solana.StringParameterDefinition{}, params[1])
		assert.IsType(t, solana.BytesParameterDefinition{}, params[2])
	})

	t.Run("TrimsWhitespaceAroundTypes", func(t *testing.T) {
		eventName, params, err := parseAnchorSignature("Transfer( uint16 , string , bytes )")

		require.NoError(t, err)
		assert.Equal(t, "Transfer", eventName)
		require.Len(t, params, 3)
		assert.IsType(t, solana.UintParameterDefinition{}, params[0])
		assert.IsType(t, solana.StringParameterDefinition{}, params[1])
		assert.IsType(t, solana.BytesParameterDefinition{}, params[2])
	})

	t.Run("EmptyParamsList", func(t *testing.T) {
		eventName, params, err := parseAnchorSignature("Ping()")

		require.NoError(t, err)
		assert.Equal(t, "Ping", eventName)
		assert.Empty(t, params)
	})

	t.Run("MissingParensIsAnError", func(t *testing.T) {
		_, _, err := parseAnchorSignature("Ping")

		assert.ErrorContains(t, err, "invalid signature")
	})

	t.Run("UnsupportedTypeIsAnError", func(t *testing.T) {
		_, _, err := parseAnchorSignature("Transfer(foo)")

		assert.ErrorContains(t, err, `unsupported parameter type "foo"`)
	})

	t.Run("DynamicArrayType", func(t *testing.T) {
		_, params, err := parseAnchorSignature("Transfer(uint16[])")

		require.NoError(t, err)
		require.Len(t, params, 1)
		arrDef, ok := params[0].(solana.ArrayParameterDefinition)
		require.True(t, ok)
		assert.Nil(t, arrDef.Length)
		uintDef, ok := arrDef.Element.(solana.UintParameterDefinition)
		require.True(t, ok)
		assert.Equal(t, 16, uintDef.BitSize)
	})

	t.Run("FixedArrayType", func(t *testing.T) {
		_, params, err := parseAnchorSignature("Transfer(uint16[4])")

		require.NoError(t, err)
		require.Len(t, params, 1)
		arrDef, ok := params[0].(solana.ArrayParameterDefinition)
		require.True(t, ok)
		require.NotNil(t, arrDef.Length)
		assert.Equal(t, 4, *arrDef.Length)
	})

	t.Run("NestedArrayType", func(t *testing.T) {
		_, params, err := parseAnchorSignature("Transfer(uint64[][])")

		require.NoError(t, err)
		require.Len(t, params, 1)
		outer, ok := params[0].(solana.ArrayParameterDefinition)
		require.True(t, ok)
		assert.Nil(t, outer.Length)
		_, ok = outer.Element.(solana.ArrayParameterDefinition)
		assert.True(t, ok)
	})

	t.Run("TupleType", func(t *testing.T) {
		_, params, err := parseAnchorSignature("Transfer((uint64,bool),string)")

		require.NoError(t, err)
		require.Len(t, params, 2)
		structDef, ok := params[0].(solana.StructParameterDefinition)
		require.True(t, ok)
		require.Len(t, structDef.Fields, 2)
		assert.IsType(t, solana.UintParameterDefinition{}, structDef.Fields[0])
		assert.IsType(t, solana.BoolParameterDefinition{}, structDef.Fields[1])
		assert.IsType(t, solana.StringParameterDefinition{}, params[1])
	})

	t.Run("ArrayOfTupleType", func(t *testing.T) {
		_, params, err := parseAnchorSignature("Transfer((uint64,bool)[])")

		require.NoError(t, err)
		require.Len(t, params, 1)
		arrDef, ok := params[0].(solana.ArrayParameterDefinition)
		require.True(t, ok)
		assert.IsType(t, solana.StructParameterDefinition{}, arrDef.Element)
	})

	t.Run("TupleOfArrayType", func(t *testing.T) {
		_, params, err := parseAnchorSignature("Transfer((uint64[],bool))")

		require.NoError(t, err)
		require.Len(t, params, 1)
		structDef, ok := params[0].(solana.StructParameterDefinition)
		require.True(t, ok)
		require.Len(t, structDef.Fields, 2)
		assert.IsType(t, solana.ArrayParameterDefinition{}, structDef.Fields[0])
	})

	t.Run("UnbalancedBracketIsAnError", func(t *testing.T) {
		_, _, err := parseAnchorSignature("Transfer(uint16[)")

		assert.Error(t, err)
	})

	t.Run("UnbalancedParenIsAnError", func(t *testing.T) {
		_, _, err := parseAnchorSignature("Transfer((uint64,bool)")

		assert.Error(t, err)
	})

	t.Run("OptionType", func(t *testing.T) {
		_, params, err := parseAnchorSignature("Transfer(option<string>,pubkey)")

		require.NoError(t, err)
		require.Len(t, params, 2)
		optDef, ok := params[0].(solana.OptionParameterDefinition)
		require.True(t, ok)
		assert.IsType(t, solana.StringParameterDefinition{}, optDef.Inner)
		assert.IsType(t, solana.PublicKeyParameterDefinition{}, params[1])
	})
}

func TestParseAnchorParameterType(t *testing.T) {
	t.Run("Scalars", func(t *testing.T) {
		bd, err := parseAnchorParameterType("bool", 0)
		require.NoError(t, err)
		assert.IsType(t, solana.BoolParameterDefinition{}, bd)

		sd, err := parseAnchorParameterType("string", 0)
		require.NoError(t, err)
		assert.IsType(t, solana.StringParameterDefinition{}, sd)

		ad, err := parseAnchorParameterType("address", 0)
		require.NoError(t, err)
		assert.IsType(t, solana.PublicKeyParameterDefinition{}, ad)

		pkd, err := parseAnchorParameterType("pubkey", 0)
		require.NoError(t, err)
		assert.IsType(t, solana.PublicKeyParameterDefinition{}, pkd)

		byd, err := parseAnchorParameterType("bytes", 0)
		require.NoError(t, err)
		assert.IsType(t, solana.BytesParameterDefinition{}, byd)

		bfd, err := parseAnchorParameterType("bytes32", 0)
		require.NoError(t, err)
		bf, ok := bfd.(solana.BytesFixedParameterDefinition)
		require.True(t, ok)
		assert.Equal(t, 32, bf.ByteLength)

		ud, err := parseAnchorParameterType("uint16", 0)
		require.NoError(t, err)
		u, ok := ud.(solana.UintParameterDefinition)
		require.True(t, ok)
		assert.Equal(t, 16, u.BitSize)

		id, err := parseAnchorParameterType("int64", 0)
		require.NoError(t, err)
		in, ok := id.(solana.IntParameterDefinition)
		require.True(t, ok)
		assert.Equal(t, 64, in.BitSize)

		fd, err := parseAnchorParameterType("float32", 0)
		require.NoError(t, err)
		f, ok := fd.(solana.FloatParameterDefinition)
		require.True(t, ok)
		assert.Equal(t, 32, f.BitSize)
	})

	t.Run("OptionOfScalar", func(t *testing.T) {
		def, err := parseAnchorParameterType("option<string>", 0)
		require.NoError(t, err)
		optDef, ok := def.(solana.OptionParameterDefinition)
		require.True(t, ok)
		assert.IsType(t, solana.StringParameterDefinition{}, optDef.Inner)
	})

	t.Run("OptionOfArray", func(t *testing.T) {
		def, err := parseAnchorParameterType("option<uint64[]>", 0)
		require.NoError(t, err)
		optDef, ok := def.(solana.OptionParameterDefinition)
		require.True(t, ok)
		assert.IsType(t, solana.ArrayParameterDefinition{}, optDef.Inner)
	})

	t.Run("OptionOfTuple", func(t *testing.T) {
		def, err := parseAnchorParameterType("option<(uint64,bool)>", 0)
		require.NoError(t, err)
		optDef, ok := def.(solana.OptionParameterDefinition)
		require.True(t, ok)
		assert.IsType(t, solana.StructParameterDefinition{}, optDef.Inner)
	})

	t.Run("NestedOption", func(t *testing.T) {
		def, err := parseAnchorParameterType("option<option<string>>", 0)
		require.NoError(t, err)
		outer, ok := def.(solana.OptionParameterDefinition)
		require.True(t, ok)
		inner, ok := outer.Inner.(solana.OptionParameterDefinition)
		require.True(t, ok)
		assert.IsType(t, solana.StringParameterDefinition{}, inner.Inner)
	})

	t.Run("OptionOfInvalidTypeIsAnError", func(t *testing.T) {
		_, err := parseAnchorParameterType("option<foo>", 0)
		assert.ErrorContains(t, err, `unsupported parameter type "foo"`)
	})

	t.Run("EmptyOptionIsAnError", func(t *testing.T) {
		_, err := parseAnchorParameterType("option<>", 0)
		assert.Error(t, err)
	})

	invalid := []string{
		"bytes0", "bytesabc", "bytes-1",
		"uint", "uintabc", "uint256", // width now required and must be a valid Solana bit size
		"int", "intabc",
		"float", "floatabc",
		"foo", "",
	}
	for _, token := range invalid {
		t.Run("Invalid_"+token, func(t *testing.T) {
			_, err := parseAnchorParameterType(token, 0)
			assert.Error(t, err)
		})
	}
}

func TestSplitTopLevel(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		tokens, err := splitTopLevel("uint64,bool,string")
		require.NoError(t, err)
		assert.Equal(t, []string{"uint64", "bool", "string"}, tokens)
	})

	t.Run("NestedTupleNotSplit", func(t *testing.T) {
		tokens, err := splitTopLevel("(uint64,bool),string")
		require.NoError(t, err)
		assert.Equal(t, []string{"(uint64,bool)", "string"}, tokens)
	})

	t.Run("NestedArrayNotSplit", func(t *testing.T) {
		tokens, err := splitTopLevel("uint64[4],bool")
		require.NoError(t, err)
		assert.Equal(t, []string{"uint64[4]", "bool"}, tokens)
	})

	t.Run("UnbalancedIsAnError", func(t *testing.T) {
		_, err := splitTopLevel("(uint64,bool")
		assert.Error(t, err)
	})

	t.Run("ExtraClosingIsAnError", func(t *testing.T) {
		_, err := splitTopLevel("uint64)")
		assert.Error(t, err)
	})
}
