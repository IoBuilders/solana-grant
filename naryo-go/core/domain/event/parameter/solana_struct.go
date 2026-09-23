package parameter

import "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"

// SolanaStructParameter is the Solana implementation of a named-field
// composite argument, matching a `defined` type in the emitting program's
// IDL.
type SolanaStructParameter struct {
	solanaParameterBase
	value []ContractEventParameter
}

func NewSolanaStructParameter(position int, value []ContractEventParameter) (SolanaStructParameter, error) {
	base, err := newSolanaParameterBase(position)
	if err != nil {
		return SolanaStructParameter{}, err
	}
	if len(value) == 0 {
		return SolanaStructParameter{}, domainerrors.NewEmptyFieldError("Value", "SolanaStructParameter")
	}
	return SolanaStructParameter{solanaParameterBase: base, value: value}, nil
}

func (p SolanaStructParameter) Type() Type {
	return TypeStruct
}

func (p SolanaStructParameter) Value() any {
	return p.value
}

var _ ContractEventParameter = SolanaStructParameter{}
