package parameter

// SolanaStringParameter is the Solana implementation of a string argument
// (Anchor's string IDL type).
type SolanaStringParameter struct {
	solanaParameterBase
	value string
}

func NewSolanaStringParameter(position int, value string) (SolanaStringParameter, error) {
	base, err := newSolanaParameterBase(position)
	if err != nil {
		return SolanaStringParameter{}, err
	}
	return SolanaStringParameter{solanaParameterBase: base, value: value}, nil
}

func (p SolanaStringParameter) Type() Type {
	return TypeString
}

func (p SolanaStringParameter) Value() any {
	return p.value
}

var _ ContractEventParameter = SolanaStringParameter{}
