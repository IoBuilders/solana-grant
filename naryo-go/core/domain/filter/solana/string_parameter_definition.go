package solana

import "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"

// StringParameterDefinition declares a string-typed field (Anchor's string
// IDL type) — a Borsh u32 length prefix followed by UTF-8 bytes, no further
// decode metadata needed.
type StringParameterDefinition struct {
	parameterDefinitionBase
}

func NewStringParameterDefinition(position int) (StringParameterDefinition, error) {
	base, err := newParameterDefinitionBase(position)
	if err != nil {
		return StringParameterDefinition{}, err
	}
	return StringParameterDefinition{parameterDefinitionBase: base}, nil
}

func (d StringParameterDefinition) Type() parameter.Type {
	return parameter.TypeString
}

var _ ParameterDefinition = StringParameterDefinition{}
