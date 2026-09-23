package solanadecoder

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/big"

	"github.com/mr-tron/base58"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter/solana"
)

// DecodeBuffer reads defs sequentially, in order, from data — a
// little-endian, unpadded Borsh buffer — and returns one
// ContractEventParameter per definition. It returns a typed buffer-underflow
// error (domainerrors.ErrDecoding) if data runs out of bytes before defs
// does, at any nesting depth.
func DecodeBuffer(data []byte, defs []solana.ParameterDefinition) ([]parameter.ContractEventParameter, error) {
	params := make([]parameter.ContractEventParameter, 0, len(defs))
	offset := 0
	for _, def := range defs {
		p, n, err := decodeValue(data[offset:], def, def.Position())
		if err != nil {
			return nil, err
		}
		params = append(params, p)
		offset += n
	}
	return params, nil
}

// requireBytes returns a buffer-underflow error if fewer than width bytes
// remain in data.
func requireBytes(data []byte, position, width int) error {
	if len(data) < width {
		return domainerrors.NewBufferUnderflowError(position, width, len(data))
	}
	return nil
}

// decodeValue decodes exactly one value described by def from the front of
// data, returning the decoded parameter and how many bytes were consumed.
func decodeValue(data []byte, def solana.ParameterDefinition, position int) (parameter.ContractEventParameter, int, error) {
	switch d := def.(type) {
	case solana.BoolParameterDefinition:
		if err := requireBytes(data, position, 1); err != nil {
			return nil, 0, err
		}
		p, err := parameter.NewSolanaBoolParameter(position, data[0] != 0)
		return p, 1, err

	case solana.UintParameterDefinition:
		width := d.BitSize / 8
		if err := requireBytes(data, position, width); err != nil {
			return nil, 0, err
		}
		p, err := parameter.NewSolanaUintParameter(position, decodeUintLE(data[:width]), d.BitSize)
		return p, width, err

	case solana.IntParameterDefinition:
		width := d.BitSize / 8
		if err := requireBytes(data, position, width); err != nil {
			return nil, 0, err
		}
		p, err := parameter.NewSolanaIntParameter(position, decodeIntLE(data[:width]), d.BitSize)
		return p, width, err

	case solana.FloatParameterDefinition:
		width := d.BitSize / 8
		if err := requireBytes(data, position, width); err != nil {
			return nil, 0, err
		}
		var f float64
		if d.BitSize == 32 {
			f = float64(math.Float32frombits(binary.LittleEndian.Uint32(data[:4])))
		} else {
			f = math.Float64frombits(binary.LittleEndian.Uint64(data[:8]))
		}
		p, err := parameter.NewSolanaFloatParameter(position, f, d.BitSize)
		return p, width, err

	case solana.PublicKeyParameterDefinition:
		if err := requireBytes(data, position, 32); err != nil {
			return nil, 0, err
		}
		p, err := parameter.NewSolanaPublicKeyParameter(position, base58.Encode(data[:32]))
		return p, 32, err

	case solana.BytesFixedParameterDefinition:
		if err := requireBytes(data, position, d.ByteLength); err != nil {
			return nil, 0, err
		}
		raw := make([]byte, d.ByteLength)
		copy(raw, data[:d.ByteLength])
		p, err := parameter.NewSolanaBytesFixedParameter(position, raw, d.ByteLength)
		return p, d.ByteLength, err

	case solana.StringParameterDefinition:
		raw, consumed, err := decodeLengthPrefixedBytes(data, position)
		if err != nil {
			return nil, 0, err
		}
		p, err := parameter.NewSolanaStringParameter(position, string(raw))
		return p, consumed, err

	case solana.BytesParameterDefinition:
		raw, consumed, err := decodeLengthPrefixedBytes(data, position)
		if err != nil {
			return nil, 0, err
		}
		p, err := parameter.NewSolanaBytesParameter(position, raw)
		return p, consumed, err

	case solana.ArrayParameterDefinition:
		return decodeArrayValue(data, d, position)

	case solana.StructParameterDefinition:
		return decodeStructValue(data, d, position)

	case solana.OptionParameterDefinition:
		return decodeOptionValue(data, d, position)
	}

	return nil, 0, domainerrors.NewInvalidFieldError("Definition", "DecodeBuffer", fmt.Sprintf("unsupported parameter definition %T", def))
}

// decodeLengthPrefixedBytes reads a Borsh u32 LE length prefix followed by
// that many raw bytes, returning the raw bytes and total bytes consumed
// (4 + length). The declared length is checked against what remains before
// slicing, so a corrupt or adversarial prefix cannot cause a slice-bounds
// panic.
func decodeLengthPrefixedBytes(data []byte, position int) ([]byte, int, error) {
	if err := requireBytes(data, position, 4); err != nil {
		return nil, 0, err
	}
	length := int(binary.LittleEndian.Uint32(data[:4]))
	if err := requireBytes(data[4:], position, length); err != nil {
		return nil, 0, err
	}
	raw := make([]byte, length)
	copy(raw, data[4:4+length])
	return raw, 4 + length, nil
}

// decodeArrayValue decodes a fixed [T; N] (d.Length set) or dynamic Vec<T>
// (d.Length nil, preceded by a u32 LE element count) by decoding Element
// that many times in sequence.
func decodeArrayValue(data []byte, d solana.ArrayParameterDefinition, position int) (parameter.ContractEventParameter, int, error) {
	offset := 0
	count := 0
	if d.Length != nil {
		count = *d.Length
	} else {
		if err := requireBytes(data, position, 4); err != nil {
			return nil, 0, err
		}
		count = int(binary.LittleEndian.Uint32(data[:4]))
		offset = 4
		// Every element needs at least 1 byte, so a corrupt/adversarial
		// count can never legitimately exceed the remaining buffer length —
		// reject it now rather than looping/allocating on a huge value.
		if err := requireBytes(data[offset:], position, count); err != nil {
			return nil, 0, err
		}
	}

	elements := make([]parameter.ContractEventParameter, 0, count)
	for i := 0; i < count; i++ {
		elemP, n, err := decodeValue(data[offset:], d.Element, i)
		if err != nil {
			// Wrap with the array's own position: decodeValue reports errors
			// against i (the element's index within the array), which reads
			// as meaningless on its own without also knowing which array
			// field, out of possibly several, failed.
			return nil, 0, fmt.Errorf("array at position %d, element %d: %w", position, i, err)
		}
		elements = append(elements, elemP)
		offset += n
	}

	p, err := parameter.NewSolanaArrayParameter(position, elements)
	return p, offset, err
}

// decodeStructValue decodes each of d.Fields in declaration order.
func decodeStructValue(data []byte, d solana.StructParameterDefinition, position int) (parameter.ContractEventParameter, int, error) {
	offset := 0
	fields := make([]parameter.ContractEventParameter, 0, len(d.Fields))
	for _, fieldDef := range d.Fields {
		fp, n, err := decodeValue(data[offset:], fieldDef, fieldDef.Position())
		if err != nil {
			return nil, 0, err
		}
		fields = append(fields, fp)
		offset += n
	}
	p, err := parameter.NewSolanaStructParameter(position, fields)
	return p, offset, err
}

// decodeOptionValue decodes a 1-byte tag (0 = None, 1 = Some) followed by
// d.Inner's encoding when the tag is 1.
func decodeOptionValue(data []byte, d solana.OptionParameterDefinition, position int) (parameter.ContractEventParameter, int, error) {
	if err := requireBytes(data, position, 1); err != nil {
		return nil, 0, err
	}
	switch data[0] {
	case 0:
		p, err := parameter.NewSolanaOptionParameter(position, nil)
		return p, 1, err
	case 1:
		innerP, n, err := decodeValue(data[1:], d.Inner, d.Inner.Position())
		if err != nil {
			return nil, 0, err
		}
		p, err := parameter.NewSolanaOptionParameter(position, innerP)
		return p, 1 + n, err
	default:
		return nil, 0, domainerrors.NewInvalidFieldError("Tag", "OptionParameterDefinition", fmt.Sprintf("invalid option tag byte %d, must be 0 or 1", data[0]))
	}
}

// decodeUintLE reads a little-endian unsigned integer of any byte length
// into a *big.Int (big.Int.SetBytes expects big-endian input, so the bytes
// are reversed first).
func decodeUintLE(raw []byte) *big.Int {
	be := make([]byte, len(raw))
	for i, b := range raw {
		be[len(raw)-1-i] = b
	}
	return new(big.Int).SetBytes(be)
}

// decodeIntLE reads a little-endian two's-complement signed integer of any
// byte length into a *big.Int.
func decodeIntLE(raw []byte) *big.Int {
	u := decodeUintLE(raw)
	bitSize := len(raw) * 8
	if u.Bit(bitSize-1) == 1 { // sign bit set -> negative
		u.Sub(u, new(big.Int).Lsh(big.NewInt(1), uint(bitSize)))
	}
	return u
}
