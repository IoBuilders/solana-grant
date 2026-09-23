package parameter

import (
	"math/big"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

// SolanaUintParameter is the Solana implementation of an unsigned integer
// argument (Anchor's u8/u16/u32/u64/u128 IDL types).
type SolanaUintParameter struct {
	solanaParameterBase
	value *big.Int

	BitSize int
}

func NewSolanaUintParameter(position int, value *big.Int, bitSize int) (SolanaUintParameter, error) {
	base, err := newSolanaParameterBase(position)
	if err != nil {
		return SolanaUintParameter{}, err
	}
	if value == nil {
		return SolanaUintParameter{}, domainerrors.NewEmptyFieldError("Value", "SolanaUintParameter")
	}
	if value.Sign() < 0 {
		return SolanaUintParameter{}, domainerrors.NewInvalidFieldError("Value", "SolanaUintParameter", "cannot be negative")
	}
	if !isValidSolanaIntegerBitSize(bitSize) {
		return SolanaUintParameter{}, domainerrors.NewInvalidFieldError("BitSize", "SolanaUintParameter", "must be one of 8, 16, 32, 64, 128")
	}
	return SolanaUintParameter{solanaParameterBase: base, value: value, BitSize: bitSize}, nil
}

func (p SolanaUintParameter) Type() Type {
	return TypeUint
}

func (p SolanaUintParameter) Value() any {
	return p.value
}

var _ ContractEventParameter = SolanaUintParameter{}
