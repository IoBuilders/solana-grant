package solana

import "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"

// BytesParameterDefinition declares a dynamically-sized byte array field
// (Anchor's bytes IDL type) — a Borsh u32 length prefix followed by raw
// bytes, no further decode metadata needed.
type BytesParameterDefinition struct {
	parameterDefinitionBase
}

func NewBytesParameterDefinition(position int) (BytesParameterDefinition, error) {
	base, err := newParameterDefinitionBase(position)
	if err != nil {
		return BytesParameterDefinition{}, err
	}
	return BytesParameterDefinition{parameterDefinitionBase: base}, nil
}

func (d BytesParameterDefinition) Type() parameter.Type {
	return parameter.TypeBytes
}

var _ ParameterDefinition = BytesParameterDefinition{}
