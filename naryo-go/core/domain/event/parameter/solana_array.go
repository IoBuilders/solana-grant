package parameter

import "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"

// SolanaArrayParameter is the Solana implementation of a homogeneous list
// argument, covering both Anchor's fixed-size arrays ([T; N]) and dynamic
// vectors (Vec<T>) IDL types. An empty (but non-nil) value is a valid,
// legitimately-decoded empty Vec<T> — only a nil value is rejected.
type SolanaArrayParameter struct {
	solanaParameterBase
	value []ContractEventParameter
}

func NewSolanaArrayParameter(position int, value []ContractEventParameter) (SolanaArrayParameter, error) {
	base, err := newSolanaParameterBase(position)
	if err != nil {
		return SolanaArrayParameter{}, err
	}
	if value == nil {
		return SolanaArrayParameter{}, domainerrors.NewEmptyFieldError("Value", "SolanaArrayParameter")
	}
	return SolanaArrayParameter{solanaParameterBase: base, value: value}, nil
}

func (p SolanaArrayParameter) Type() Type {
	return TypeArray
}

func (p SolanaArrayParameter) Value() any {
	return p.value
}

var _ ContractEventParameter = SolanaArrayParameter{}
