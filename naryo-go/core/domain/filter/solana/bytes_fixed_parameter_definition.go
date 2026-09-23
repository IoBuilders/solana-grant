package solana

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
)

// BytesFixedParameterDefinition declares a fixed-length byte array field
// (e.g. Anchor's [u8; N] IDL arrays or a 32-byte hash) — exactly ByteLength
// raw bytes, no length prefix.
type BytesFixedParameterDefinition struct {
	parameterDefinitionBase

	ByteLength int
}

func NewBytesFixedParameterDefinition(position, byteLength int) (BytesFixedParameterDefinition, error) {
	base, err := newParameterDefinitionBase(position)
	if err != nil {
		return BytesFixedParameterDefinition{}, err
	}
	if byteLength < 1 {
		return BytesFixedParameterDefinition{}, domainerrors.NewInvalidFieldError("ByteLength", "BytesFixedParameterDefinition", "must be at least 1")
	}
	return BytesFixedParameterDefinition{parameterDefinitionBase: base, ByteLength: byteLength}, nil
}

func (d BytesFixedParameterDefinition) Type() parameter.Type {
	return parameter.TypeBytesFixed
}

var _ ParameterDefinition = BytesFixedParameterDefinition{}
