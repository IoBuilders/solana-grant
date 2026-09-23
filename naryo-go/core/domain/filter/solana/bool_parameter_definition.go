package solana

import "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"

// BoolParameterDefinition declares a boolean-typed field (Anchor's bool IDL
// type) — a single byte, no further decode metadata needed.
type BoolParameterDefinition struct {
	parameterDefinitionBase
}

func NewBoolParameterDefinition(position int) (BoolParameterDefinition, error) {
	base, err := newParameterDefinitionBase(position)
	if err != nil {
		return BoolParameterDefinition{}, err
	}
	return BoolParameterDefinition{parameterDefinitionBase: base}, nil
}

func (d BoolParameterDefinition) Type() parameter.Type {
	return parameter.TypeBool
}

var _ ParameterDefinition = BoolParameterDefinition{}
