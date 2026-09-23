package parameter

import (
	"math/big"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

// SolanaIntParameter is the Solana implementation of a signed integer
// argument (Anchor's i8/i16/i32/i64/i128 IDL types).
type SolanaIntParameter struct {
	solanaParameterBase
	value *big.Int

	BitSize int
}

func NewSolanaIntParameter(position int, value *big.Int, bitSize int) (SolanaIntParameter, error) {
	base, err := newSolanaParameterBase(position)
	if err != nil {
		return SolanaIntParameter{}, err
	}
	if value == nil {
		return SolanaIntParameter{}, domainerrors.NewEmptyFieldError("Value", "SolanaIntParameter")
	}
	if !isValidSolanaIntegerBitSize(bitSize) {
		return SolanaIntParameter{}, domainerrors.NewInvalidFieldError("BitSize", "SolanaIntParameter", "must be one of 8, 16, 32, 64, 128")
	}
	return SolanaIntParameter{solanaParameterBase: base, value: value, BitSize: bitSize}, nil
}

func (p SolanaIntParameter) Type() Type {
	return TypeInt
}

func (p SolanaIntParameter) Value() any {
	return p.value
}

var _ ContractEventParameter = SolanaIntParameter{}
