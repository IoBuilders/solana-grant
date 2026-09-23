package solana

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
)

// ArrayParameterDefinition declares a homogeneous list field, covering both
// Anchor's fixed-size arrays ([T; N]) and dynamic vectors (Vec<T>) IDL
// types. Length is nil for a dynamic Vec<T> (Borsh: a u32 length prefix
// precedes the elements) and non-nil for a fixed [T; N] (Borsh: exactly
// *Length elements, no prefix).
type ArrayParameterDefinition struct {
	parameterDefinitionBase

	Element ParameterDefinition
	Length  *int
}

func NewArrayParameterDefinition(position int, element ParameterDefinition, length *int) (ArrayParameterDefinition, error) {
	base, err := newParameterDefinitionBase(position)
	if err != nil {
		return ArrayParameterDefinition{}, err
	}
	if element == nil {
		return ArrayParameterDefinition{}, domainerrors.NewEmptyFieldError("Element", "ArrayParameterDefinition")
	}
	if length != nil && *length < 1 {
		return ArrayParameterDefinition{}, domainerrors.NewInvalidFieldError("Length", "ArrayParameterDefinition", "must be at least 1 when set")
	}
	return ArrayParameterDefinition{parameterDefinitionBase: base, Element: element, Length: length}, nil
}

func (d ArrayParameterDefinition) Type() parameter.Type {
	return parameter.TypeArray
}

var _ ParameterDefinition = ArrayParameterDefinition{}
