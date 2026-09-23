package solana

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
)

// IntParameterDefinition declares a signed-integer-typed field (Anchor's
// i8/i16/i32/i64/i128 IDL types). BitSize drives how many little-endian,
// two's-complement bytes a decoder must read.
type IntParameterDefinition struct {
	parameterDefinitionBase

	BitSize int
}

func NewIntParameterDefinition(position, bitSize int) (IntParameterDefinition, error) {
	base, err := newParameterDefinitionBase(position)
	if err != nil {
		return IntParameterDefinition{}, err
	}
	if !isValidIntegerBitSize(bitSize) {
		return IntParameterDefinition{}, domainerrors.NewInvalidFieldError("BitSize", "IntParameterDefinition", "must be one of 8, 16, 32, 64, 128")
	}
	return IntParameterDefinition{parameterDefinitionBase: base, BitSize: bitSize}, nil
}

func (d IntParameterDefinition) Type() parameter.Type {
	return parameter.TypeInt
}

var _ ParameterDefinition = IntParameterDefinition{}
