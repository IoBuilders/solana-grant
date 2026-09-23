package solana

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
)

// UintParameterDefinition declares an unsigned-integer-typed field (Anchor's
// u8/u16/u32/u64/u128 IDL types). BitSize drives how many little-endian
// bytes a decoder must read.
type UintParameterDefinition struct {
	parameterDefinitionBase

	BitSize int
}

func NewUintParameterDefinition(position, bitSize int) (UintParameterDefinition, error) {
	base, err := newParameterDefinitionBase(position)
	if err != nil {
		return UintParameterDefinition{}, err
	}
	if !isValidIntegerBitSize(bitSize) {
		return UintParameterDefinition{}, domainerrors.NewInvalidFieldError("BitSize", "UintParameterDefinition", "must be one of 8, 16, 32, 64, 128")
	}
	return UintParameterDefinition{parameterDefinitionBase: base, BitSize: bitSize}, nil
}

func (d UintParameterDefinition) Type() parameter.Type {
	return parameter.TypeUint
}

var _ ParameterDefinition = UintParameterDefinition{}
