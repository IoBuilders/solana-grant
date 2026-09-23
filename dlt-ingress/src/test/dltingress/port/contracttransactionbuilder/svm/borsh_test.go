package svmidl

import (
	"bytes"
	"dlt-ingress/src/main/dltingress/port/contracttransactionbuilder/svm"
	"math"
	"math/big"
	"reflect"
	"testing"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func encodeArg(t *testing.T, kind svmidl.Kind, v any) ([]byte, error) {
	t.Helper()
	var buf bytes.Buffer
	enc := bin.NewBorshEncoder(&buf)
	err := svmidl.BorshEncode(svmidl.IdlType{Kind: kind}, v, enc, nil)
	return buf.Bytes(), err
}

func TestDltIngressBorsh_Primitives_HappyPath(t *testing.T) {
	pk := solana.MustPublicKeyFromBase58("11111111111111111111111111111111")

	cases := []struct {
		name string
		kind svmidl.Kind
		in   any
		out  []byte
	}{
		{"bool true", svmidl.KindBool, true, []byte{0x01}},
		{"bool false", svmidl.KindBool, false, []byte{0x00}},
		{"u8", svmidl.KindU8, uint64(0xAB), []byte{0xAB}},
		{"u16", svmidl.KindU16, uint64(0x1234), []byte{0x34, 0x12}},
		{"u32", svmidl.KindU32, uint64(0xDEADBEEF), []byte{0xEF, 0xBE, 0xAD, 0xDE}},
		{"u64", svmidl.KindU64, uint64(1), []byte{0x01, 0, 0, 0, 0, 0, 0, 0}},
		{"i8 negative", svmidl.KindI8, int64(-1), []byte{0xFF}},
		{"i16 negative", svmidl.KindI16, int64(-2), []byte{0xFE, 0xFF}},
		{"i32 negative", svmidl.KindI32, int64(-1), []byte{0xFF, 0xFF, 0xFF, 0xFF}},
		{"i64 negative", svmidl.KindI64, int64(-1),
			[]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}},
		{"f32", svmidl.KindF32, float32(1.0), []byte{0x00, 0x00, 0x80, 0x3F}},
		{"f64", svmidl.KindF64, float64(1.0),
			[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF0, 0x3F}},
		{"string", svmidl.KindString, "ab", []byte{0x02, 0x00, 0x00, 0x00, 'a', 'b'}},
		{"pubkey (32 raw)", svmidl.KindPubkey, pk, pk[:]},
		{"bytes", svmidl.KindBytes, []byte{0xCA, 0xFE},
			[]byte{0x02, 0x00, 0x00, 0x00, 0xCA, 0xFE}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := encodeArg(t, c.kind, c.in)
			require.NoError(t, err)
			if !reflect.DeepEqual(got, c.out) {
				t.Fatalf("got %x, want %x", got, c.out)
			}
		})
	}
}

func TestDltIngressBorsh_U128_LittleEndian(t *testing.T) {
	got, err := encodeArg(t, svmidl.KindU128, big.NewInt(1))
	require.NoError(t, err)
	want := append([]byte{0x01}, make([]byte, 15)...)
	assert.Equal(t, want, got)

	huge := new(big.Int).Lsh(big.NewInt(1), 64) // 2^64
	got, err = encodeArg(t, svmidl.KindU128, huge)
	require.NoError(t, err)
	want = append(make([]byte, 8), append([]byte{0x01}, make([]byte, 7)...)...)
	assert.Equal(t, want, got)
}

func TestDltIngressBorsh_I128_TwosComplement(t *testing.T) {
	got, err := encodeArg(t, svmidl.KindI128, big.NewInt(-1))
	require.NoError(t, err)
	assert.Equal(t, bytes.Repeat([]byte{0xFF}, 16), got)

	got, err = encodeArg(t, svmidl.KindI128, big.NewInt(1))
	require.NoError(t, err)
	assert.Equal(t, append([]byte{0x01}, make([]byte, 15)...), got)
}

func TestDltIngressBorsh_OverflowDetected(t *testing.T) {
	_, err := encodeArg(t, svmidl.KindU8, uint64(math.MaxUint8)+1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "overflows u8")

	_, err = encodeArg(t, svmidl.KindI8, int64(math.MaxInt8)+1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "overflows i8")

	_, err = encodeArg(t, svmidl.KindU128, big.NewInt(-1))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "u128: negative value not allowed")
}

func TestDltIngressBorsh_TypeMismatchDetected(t *testing.T) {
	_, err := encodeArg(t, svmidl.KindU32, 42)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "incompatible Go type")

	_, err = encodeArg(t, svmidl.KindString, []byte("ab"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "kind \"string\"")
}

func TestDltIngressBorsh_UnknownKindRejected(t *testing.T) {
	_, err := encodeArg(t, svmidl.Kind("madeUp"), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported IDL kind")
}
