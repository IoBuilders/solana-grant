package parameter

// SolanaBoolParameter is the Solana implementation of a boolean argument
// (Anchor's bool IDL type).
type SolanaBoolParameter struct {
	solanaParameterBase
	value bool
}

func NewSolanaBoolParameter(position int, value bool) (SolanaBoolParameter, error) {
	base, err := newSolanaParameterBase(position)
	if err != nil {
		return SolanaBoolParameter{}, err
	}
	return SolanaBoolParameter{solanaParameterBase: base, value: value}, nil
}

func (p SolanaBoolParameter) Type() Type {
	return TypeBool
}

func (p SolanaBoolParameter) Value() any {
	return p.value
}

var _ ContractEventParameter = SolanaBoolParameter{}
