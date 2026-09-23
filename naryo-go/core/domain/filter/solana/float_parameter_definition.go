package solana

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
)

// FloatParameterDefinition declares a floating-point-typed field (Anchor's
// f32/f64 IDL types). BitSize drives how many little-endian IEEE-754 bytes
// a decoder must read.
type FloatParameterDefinition struct {
	parameterDefinitionBase

	BitSize int
}

func NewFloatParameterDefinition(position, bitSize int) (FloatParameterDefinition, error) {
	base, err := newParameterDefinitionBase(position)
	if err != nil {
		return FloatParameterDefinition{}, err
	}
	if bitSize != 32 && bitSize != 64 {
		return FloatParameterDefinition{}, domainerrors.NewInvalidFieldError("BitSize", "FloatParameterDefinition", "must be 32 or 64")
	}
	return FloatParameterDefinition{parameterDefinitionBase: base, BitSize: bitSize}, nil
}

func (d FloatParameterDefinition) Type() parameter.Type {
	return parameter.TypeFloat
}

var _ ParameterDefinition = FloatParameterDefinition{}
