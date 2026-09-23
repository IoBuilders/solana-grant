package parameter

import "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"

// SolanaBytesFixedParameter is the Solana implementation of a fixed-length
// byte array argument, e.g. Anchor's [u8; N] IDL arrays or a 32-byte hash.
type SolanaBytesFixedParameter struct {
	solanaParameterBase
	value []byte

	ByteLength int
}

func NewSolanaBytesFixedParameter(position int, value []byte, byteLength int) (SolanaBytesFixedParameter, error) {
	base, err := newSolanaParameterBase(position)
	if err != nil {
		return SolanaBytesFixedParameter{}, err
	}
	if byteLength < 1 {
		return SolanaBytesFixedParameter{}, domainerrors.NewInvalidFieldError("ByteLength", "SolanaBytesFixedParameter", "must be at least 1")
	}
	if len(value) != byteLength {
		return SolanaBytesFixedParameter{}, domainerrors.NewInvalidFieldError("Value", "SolanaBytesFixedParameter", "length does not match ByteLength")
	}
	return SolanaBytesFixedParameter{solanaParameterBase: base, value: value, ByteLength: byteLength}, nil
}

func (p SolanaBytesFixedParameter) Type() Type {
	return TypeBytesFixed
}

func (p SolanaBytesFixedParameter) Value() any {
	return p.value
}

var _ ContractEventParameter = SolanaBytesFixedParameter{}
