package parameter

import "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"

// SolanaFloatParameter is the Solana implementation of a floating point
// argument (Anchor's f32/f64 IDL types).
type SolanaFloatParameter struct {
	solanaParameterBase
	value float64

	BitSize int
}

func NewSolanaFloatParameter(position int, value float64, bitSize int) (SolanaFloatParameter, error) {
	base, err := newSolanaParameterBase(position)
	if err != nil {
		return SolanaFloatParameter{}, err
	}
	if bitSize != 32 && bitSize != 64 {
		return SolanaFloatParameter{}, domainerrors.NewInvalidFieldError("BitSize", "SolanaFloatParameter", "must be 32 or 64")
	}
	return SolanaFloatParameter{solanaParameterBase: base, value: value, BitSize: bitSize}, nil
}

func (p SolanaFloatParameter) Type() Type {
	return TypeFloat
}

func (p SolanaFloatParameter) Value() any {
	return p.value
}

var _ ContractEventParameter = SolanaFloatParameter{}
