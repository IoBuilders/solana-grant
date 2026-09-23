package parameter

import "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"

// SolanaBytesParameter is the Solana implementation of a dynamically-sized
// byte array argument (Anchor's bytes IDL type).
type SolanaBytesParameter struct {
	solanaParameterBase
	value []byte
}

func NewSolanaBytesParameter(position int, value []byte) (SolanaBytesParameter, error) {
	base, err := newSolanaParameterBase(position)
	if err != nil {
		return SolanaBytesParameter{}, err
	}
	if value == nil {
		return SolanaBytesParameter{}, domainerrors.NewEmptyFieldError("Value", "SolanaBytesParameter")
	}
	return SolanaBytesParameter{solanaParameterBase: base, value: value}, nil
}

func (p SolanaBytesParameter) Type() Type {
	return TypeBytes
}

func (p SolanaBytesParameter) Value() any {
	return p.value
}

var _ ContractEventParameter = SolanaBytesParameter{}
