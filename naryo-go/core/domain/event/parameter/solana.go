package parameter

import "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"

// solanaParameterBase holds the fields and validation common to every
// Solana ContractEventParameter implementation.
type solanaParameterBase struct {
	position int
}

func newSolanaParameterBase(position int) (solanaParameterBase, error) {
	if position < 0 {
		return solanaParameterBase{}, domainerrors.NewInvalidFieldError("Position", "ContractEventParameter", "must be non-negative")
	}
	return solanaParameterBase{position: position}, nil
}

func (b solanaParameterBase) Position() int {
	return b.position
}

func isValidSolanaIntegerBitSize(bitSize int) bool {
	switch bitSize {
	case 8, 16, 32, 64, 128:
		return true
	}
	return false
}
