package svmidl

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/big"

	"github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
)

func BorshEncode(t IdlType, v any, enc *bin.Encoder, idl *Idl) error {
	switch t.Kind {
	case KindBool:
		b, ok := v.(bool)
		if !ok {
			return typeErr(t.Kind, v)
		}
		return enc.WriteBool(b)
	case KindU8:
		n, err := asUint64(v, t.Kind)
		if err != nil {
			return err
		}
		if n > math.MaxUint8 {
			return overflowErr(t.Kind, n)
		}
		return enc.WriteUint8(uint8(n))
	case KindU16:
		n, err := asUint64(v, t.Kind)
		if err != nil {
			return err
		}
		if n > math.MaxUint16 {
			return overflowErr(t.Kind, n)
		}
		return enc.WriteUint16(uint16(n), bin.LE)
	case KindU32:
		n, err := asUint64(v, t.Kind)
		if err != nil {
			return err
		}
		if n > math.MaxUint32 {
			return overflowErr(t.Kind, n)
		}
		return enc.WriteUint32(uint32(n), bin.LE)
	case KindU64:
		n, err := asUint64(v, t.Kind)
		if err != nil {
			return err
		}
		return enc.WriteUint64(n, bin.LE)
	case KindU128:
		bi, ok := v.(*big.Int)
		if !ok {
			return typeErr(t.Kind, v)
		}
		u, err := bigIntToUint128(bi)
		if err != nil {
			return err
		}
		return enc.WriteUint128(u, bin.LE)
	case KindI8:
		n, err := asInt64(v, t.Kind)
		if err != nil {
			return err
		}
		if n < math.MinInt8 || n > math.MaxInt8 {
			return overflowErr(t.Kind, n)
		}
		return enc.WriteInt8(int8(n))
	case KindI16:
		n, err := asInt64(v, t.Kind)
		if err != nil {
			return err
		}
		if n < math.MinInt16 || n > math.MaxInt16 {
			return overflowErr(t.Kind, n)
		}
		return enc.WriteInt16(int16(n), bin.LE)
	case KindI32:
		n, err := asInt64(v, t.Kind)
		if err != nil {
			return err
		}
		if n < math.MinInt32 || n > math.MaxInt32 {
			return overflowErr(t.Kind, n)
		}
		return enc.WriteInt32(int32(n), bin.LE)
	case KindI64:
		n, err := asInt64(v, t.Kind)
		if err != nil {
			return err
		}
		return enc.WriteInt64(n, bin.LE)
	case KindI128:
		bi, ok := v.(*big.Int)
		if !ok {
			return typeErr(t.Kind, v)
		}
		i, err := bigIntToInt128(bi)
		if err != nil {
			return err
		}
		return enc.WriteInt128(i, bin.LE)
	case KindF32:
		f, ok := v.(float32)
		if !ok {
			return typeErr(t.Kind, v)
		}
		return enc.WriteFloat32(f, bin.LE)
	case KindF64:
		f, ok := v.(float64)
		if !ok {
			return typeErr(t.Kind, v)
		}
		return enc.WriteFloat64(f, bin.LE)
	case KindString:
		s, ok := v.(string)
		if !ok {
			return typeErr(t.Kind, v)
		}
		return enc.WriteString(s)
	case KindPubkey:
		pk, ok := v.(solana.PublicKey)
		if !ok {
			return typeErr(t.Kind, v)
		}
		return enc.WriteBytes(pk[:], false)
	case KindBytes:
		b, ok := v.([]byte)
		if !ok {
			return typeErr(t.Kind, v)
		}
		return enc.WriteBytes(b, true)
	case KindArray:
		return encodeArray(t, v, enc, idl)
	case KindVec:
		return encodeVec(t, v, enc, idl)
	case KindOption:
		return encodeOption(t, v, enc, idl)
	case KindStruct:
		return encodeStruct(t.Fields, v, enc, idl)
	case KindEnum:
		return encodeEnum(t.Variants, v, enc, idl)
	case KindDefined:
		td, err := idl.ResolveDefined(t.Defined)
		if err != nil {
			return err
		}
		return BorshEncode(td.Type, v, enc, idl)
	}

	return fmt.Errorf("borsh encode: unsupported IDL kind %q", t.Kind)
}

func encodeArray(t IdlType, v any, enc *bin.Encoder, idl *Idl) error {
	if t.Inner == nil {
		return fmt.Errorf("borsh encode: array missing inner type")
	}
	items, ok := v.([]any)
	if !ok {
		return typeErr(t.Kind, v)
	}
	if len(items) != t.Length {
		return fmt.Errorf("borsh encode: array expected length %d, got %d", t.Length, len(items))
	}
	for i, item := range items {
		if err := BorshEncode(*t.Inner, item, enc, idl); err != nil {
			return fmt.Errorf("array[%d]: %w", i, err)
		}
	}
	return nil
}

func encodeVec(t IdlType, v any, enc *bin.Encoder, idl *Idl) error {
	if t.Inner == nil {
		return fmt.Errorf("borsh encode: vec missing inner type")
	}
	items, ok := v.([]any)
	if !ok {
		return typeErr(t.Kind, v)
	}
	if err := enc.WriteUint32(uint32(len(items)), bin.LE); err != nil {
		return err
	}
	for i, item := range items {
		if err := BorshEncode(*t.Inner, item, enc, idl); err != nil {
			return fmt.Errorf("vec[%d]: %w", i, err)
		}
	}
	return nil
}

func encodeOption(t IdlType, v any, enc *bin.Encoder, idl *Idl) error {
	if t.Inner == nil {
		return fmt.Errorf("borsh encode: option missing inner type")
	}
	if v == nil {
		return enc.WriteByte(0)
	}
	if err := enc.WriteByte(1); err != nil {
		return err
	}
	return BorshEncode(*t.Inner, v, enc, idl)
}

func encodeStruct(fields []Field, v any, enc *bin.Encoder, idl *Idl) error {
	m, ok := v.(map[string]any)
	if !ok {
		return fmt.Errorf("borsh encode: struct expects map[string]any, got %T", v)
	}
	for _, f := range fields {
		fv, present := m[f.Name]
		if !present && f.Type.Kind != KindOption {
			return fmt.Errorf("borsh encode: struct missing field %q", f.Name)
		}
		if err := BorshEncode(f.Type, fv, enc, idl); err != nil {
			return fmt.Errorf("field %q: %w", f.Name, err)
		}
	}
	return nil
}

func encodeEnum(variants []EnumVariant, v any, enc *bin.Encoder, idl *Idl) error {
	m, ok := v.(map[string]any)
	if !ok {
		return fmt.Errorf("borsh encode: enum expects map[string]any, got %T", v)
	}
	if len(m) != 1 {
		return fmt.Errorf("borsh encode: enum expects exactly one variant key, got %d", len(m))
	}
	var name string
	var payload any
	for k, val := range m {
		name = k
		payload = val
	}
	for i, variant := range variants {
		if variant.Name != name {
			continue
		}
		if i > 255 {
			return fmt.Errorf("borsh encode: enum variant index %d exceeds u8", i)
		}
		if err := enc.WriteByte(byte(i)); err != nil {
			return err
		}
		if len(variant.Fields) == 0 {
			return nil
		}
		return encodeStruct(variant.Fields, payload, enc, idl)
	}
	return fmt.Errorf("borsh encode: enum variant %q not declared", name)
}

func asUint64(v any, k Kind) (uint64, error) {
	if n, ok := v.(uint64); ok {
		return n, nil
	}
	return 0, typeErr(k, v)
}

func asInt64(v any, k Kind) (int64, error) {
	if n, ok := v.(int64); ok {
		return n, nil
	}
	return 0, typeErr(k, v)
}

func typeErr(k Kind, v any) error {
	return fmt.Errorf("borsh encode: kind %q got incompatible Go type %T", k, v)
}

func overflowErr(k Kind, v any) error {
	return fmt.Errorf("borsh encode: value %v overflows %s", v, k)
}

func bigIntToUint128(v *big.Int) (bin.Uint128, error) {
	if v == nil {
		return bin.Uint128{}, fmt.Errorf("u128: nil *big.Int")
	}
	if v.Sign() < 0 {
		return bin.Uint128{}, fmt.Errorf("u128: negative value not allowed")
	}
	if v.BitLen() > 128 {
		return bin.Uint128{}, fmt.Errorf("u128: value exceeds 128 bits")
	}
	be := v.FillBytes(make([]byte, 16))
	return bin.Uint128{
		Hi:         binary.BigEndian.Uint64(be[:8]),
		Lo:         binary.BigEndian.Uint64(be[8:]),
		Endianness: binary.LittleEndian,
	}, nil
}

func bigIntToInt128(v *big.Int) (bin.Int128, error) {
	if v == nil {
		return bin.Int128{}, fmt.Errorf("i128: nil *big.Int")
	}
	min128 := new(big.Int).Lsh(big.NewInt(-1), 127)
	max128 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 127), big.NewInt(1))
	if v.Cmp(min128) < 0 || v.Cmp(max128) > 0 {
		return bin.Int128{}, fmt.Errorf("i128: value out of range")
	}
	be := make([]byte, 16)
	if v.Sign() >= 0 {
		v.FillBytes(be)
	} else {
		twos := new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 128), v)
		twos.FillBytes(be)
	}
	return bin.Int128{
		Hi:         binary.BigEndian.Uint64(be[:8]),
		Lo:         binary.BigEndian.Uint64(be[8:]),
		Endianness: binary.LittleEndian,
	}, nil
}
