package svmidl

import (
	"bytes"
	"dlt-ingress/src/main/dltingress/port/contracttransactionbuilder/svm"
	"testing"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func encode(t *testing.T, ty svmidl.IdlType, v any, idl *svmidl.Idl) ([]byte, error) {
	t.Helper()
	var buf bytes.Buffer
	enc := bin.NewBorshEncoder(&buf)
	err := svmidl.BorshEncode(ty, v, enc, idl)
	return buf.Bytes(), err
}

func u8Type() svmidl.IdlType {
	return svmidl.IdlType{Kind: svmidl.KindU8}
}

func TestDltIngressBorsh_Vec_U8(t *testing.T) {
	inner := u8Type()
	ty := svmidl.IdlType{Kind: svmidl.KindVec, Inner: &inner}

	got, err := encode(t, ty, []any{uint64(1), uint64(2), uint64(3)}, nil)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x03, 0, 0, 0, 0x01, 0x02, 0x03}, got)
}

func TestDltIngressBorsh_Vec_Empty(t *testing.T) {
	inner := svmidl.IdlType{Kind: svmidl.KindString}
	ty := svmidl.IdlType{Kind: svmidl.KindVec, Inner: &inner}

	got, err := encode(t, ty, []any{}, nil)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x00, 0x00, 0x00, 0x00}, got)
}

func TestDltIngressBorsh_Option_NoneSome(t *testing.T) {
	inner := svmidl.IdlType{Kind: svmidl.KindU32}
	ty := svmidl.IdlType{Kind: svmidl.KindOption, Inner: &inner}

	none, err := encode(t, ty, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x00}, none)

	some, err := encode(t, ty, uint64(1), nil)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x01, 0x01, 0x00, 0x00, 0x00}, some)
}

func TestDltIngressBorsh_Array_Fixed(t *testing.T) {
	inner := u8Type()
	ty := svmidl.IdlType{Kind: svmidl.KindArray, Inner: &inner, Length: 3}
	got, err := encode(t, ty, []any{uint64(10), uint64(20), uint64(30)}, nil)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x0A, 0x14, 0x1E}, got)
}

func TestDltIngressBorsh_Array_LengthMismatch(t *testing.T) {
	inner := u8Type()
	ty := svmidl.IdlType{Kind: svmidl.KindArray, Inner: &inner, Length: 3}
	_, err := encode(t, ty, []any{uint64(1), uint64(2)}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "array expected length 3, got 2")
}

func TestDltIngressBorsh_Struct_PreservesFieldOrder(t *testing.T) {
	ty := svmidl.IdlType{
		Kind: svmidl.KindStruct,
		Fields: []svmidl.Field{
			{Name: "a", Type: svmidl.IdlType{Kind: svmidl.KindU8}},
			{Name: "b", Type: svmidl.IdlType{Kind: svmidl.KindString}},
		},
	}
	got, err := encode(t, ty, map[string]any{"b": "ok", "a": uint64(7)}, nil)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x07, 0x02, 0, 0, 0, 'o', 'k'}, got)
}

func TestDltIngressBorsh_Struct_MissingFieldErrors(t *testing.T) {
	ty := svmidl.IdlType{
		Kind: svmidl.KindStruct,
		Fields: []svmidl.Field{
			{Name: "a", Type: svmidl.IdlType{Kind: svmidl.KindU8}},
		},
	}
	_, err := encode(t, ty, map[string]any{}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `missing field "a"`)
}

func TestDltIngressBorsh_Struct_OmittedOptionIsNone(t *testing.T) {
	inner := svmidl.IdlType{Kind: svmidl.KindU32}
	ty := svmidl.IdlType{
		Kind: svmidl.KindStruct,
		Fields: []svmidl.Field{
			{Name: "maybe", Type: svmidl.IdlType{Kind: svmidl.KindOption, Inner: &inner}},
		},
	}
	got, err := encode(t, ty, map[string]any{}, nil)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x00}, got)
}

func TestDltIngressBorsh_Enum_UnitAndPayloadVariants(t *testing.T) {
	ty := svmidl.IdlType{
		Kind: svmidl.KindEnum,
		Variants: []svmidl.EnumVariant{
			{Name: "Pending"},
			{Name: "Done", Fields: []svmidl.Field{
				{Name: "ts", Type: svmidl.IdlType{Kind: svmidl.KindI64}},
			}},
		},
	}

	pending, err := encode(t, ty, map[string]any{"Pending": nil}, nil)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x00}, pending)

	done, err := encode(t, ty, map[string]any{"Done": map[string]any{"ts": int64(1)}}, nil)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x01, 0x01, 0, 0, 0, 0, 0, 0, 0}, done)
}

func TestDltIngressBorsh_Enum_UnknownVariantErrors(t *testing.T) {
	ty := svmidl.IdlType{
		Kind: svmidl.KindEnum,
		Variants: []svmidl.EnumVariant{
			{Name: "Only"},
		},
	}
	_, err := encode(t, ty, map[string]any{"Missing": nil}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `variant "Missing" not declared`)
}

func TestDltIngressBorsh_Defined_ResolvesViaIdl(t *testing.T) {
	idl := &svmidl.Idl{
		Types: []svmidl.TypeDef{{
			Name: "Config",
			Type: svmidl.IdlType{
				Kind: svmidl.KindStruct,
				Fields: []svmidl.Field{
					{Name: "owner", Type: svmidl.IdlType{Kind: svmidl.KindPubkey}},
					{Name: "active", Type: svmidl.IdlType{Kind: svmidl.KindBool}},
				},
			},
		}},
	}
	ty := svmidl.IdlType{Kind: svmidl.KindDefined, Defined: "Config"}
	pk := solana.MustPublicKeyFromBase58("11111111111111111111111111111111")

	got, err := encode(t, ty, map[string]any{"owner": pk, "active": true}, idl)
	require.NoError(t, err)
	want := append(pk[:], 0x01)
	assert.Equal(t, want, got)
}

func TestDltIngressBorsh_Defined_MissingTypeErrors(t *testing.T) {
	ty := svmidl.IdlType{Kind: svmidl.KindDefined, Defined: "Nope"}
	_, err := encode(t, ty, nil, &svmidl.Idl{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `defined type "Nope" not found`)
}

func TestDltIngressBorsh_StructWithVecAndOption_Composition(t *testing.T) {
	str := svmidl.IdlType{Kind: svmidl.KindString}
	u32 := svmidl.IdlType{Kind: svmidl.KindU32}
	vecStr := svmidl.IdlType{Kind: svmidl.KindVec, Inner: &str}
	optU32 := svmidl.IdlType{Kind: svmidl.KindOption, Inner: &u32}

	ty := svmidl.IdlType{
		Kind: svmidl.KindStruct,
		Fields: []svmidl.Field{
			{Name: "tags", Type: vecStr},
			{Name: "note", Type: optU32},
		},
	}
	got, err := encode(t, ty, map[string]any{
		"tags": []any{"a", "bb"},
		"note": uint64(7),
	}, nil)
	require.NoError(t, err)

	expected := []byte{}
	expected = append(expected, 0x02, 0, 0, 0)
	expected = append(expected, 0x01, 0, 0, 0, 'a')
	expected = append(expected, 0x02, 0, 0, 0, 'b', 'b')
	expected = append(expected, 0x01, 0x07, 0, 0, 0)
	assert.Equal(t, expected, got)
}
