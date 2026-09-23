package solana

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
)

// StructParameterDefinition declares a named-field composite field, matching
// a `defined` type in the emitting program's IDL. Fields are purely
// positional (no field names), matching parameter.SolanaStructParameter's
// own shape on the decoded-output side.
type StructParameterDefinition struct {
	parameterDefinitionBase

	Fields []ParameterDefinition
}

func NewStructParameterDefinition(position int, fields []ParameterDefinition) (StructParameterDefinition, error) {
	base, err := newParameterDefinitionBase(position)
	if err != nil {
		return StructParameterDefinition{}, err
	}
	if len(fields) == 0 {
		return StructParameterDefinition{}, domainerrors.NewEmptyFieldError("Fields", "StructParameterDefinition")
	}
	if err := validateNoDuplicatePositions(fields, "StructParameterDefinition"); err != nil {
		return StructParameterDefinition{}, err
	}
	return StructParameterDefinition{parameterDefinitionBase: base, Fields: fields}, nil
}

func (d StructParameterDefinition) Type() parameter.Type {
	return parameter.TypeStruct
}

var _ ParameterDefinition = StructParameterDefinition{}
