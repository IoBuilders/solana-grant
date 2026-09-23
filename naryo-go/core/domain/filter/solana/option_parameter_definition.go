package solana

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
)

// OptionParameterDefinition declares a nested field that may be absent,
// matching Anchor's option<T> IDL type. Borsh: a 1-byte tag (0 = None, 1 =
// Some) followed by Inner's encoding when the tag is 1.
type OptionParameterDefinition struct {
	parameterDefinitionBase

	Inner ParameterDefinition
}

func NewOptionParameterDefinition(position int, inner ParameterDefinition) (OptionParameterDefinition, error) {
	base, err := newParameterDefinitionBase(position)
	if err != nil {
		return OptionParameterDefinition{}, err
	}
	if inner == nil {
		return OptionParameterDefinition{}, domainerrors.NewEmptyFieldError("Inner", "OptionParameterDefinition")
	}
	return OptionParameterDefinition{parameterDefinitionBase: base, Inner: inner}, nil
}

func (d OptionParameterDefinition) Type() parameter.Type {
	return parameter.TypeOption
}

var _ ParameterDefinition = OptionParameterDefinition{}
