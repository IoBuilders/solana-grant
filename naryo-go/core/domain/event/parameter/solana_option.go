package parameter

// SolanaOptionParameter is the Solana implementation of a nested parameter
// that may be absent, matching Anchor's option<T> IDL type. Value is nil
// when the IDL option is None.
type SolanaOptionParameter struct {
	solanaParameterBase
	value ContractEventParameter
}

func NewSolanaOptionParameter(position int, value ContractEventParameter) (SolanaOptionParameter, error) {
	base, err := newSolanaParameterBase(position)
	if err != nil {
		return SolanaOptionParameter{}, err
	}
	return SolanaOptionParameter{solanaParameterBase: base, value: value}, nil
}

func (p SolanaOptionParameter) Type() Type {
	return TypeOption
}

func (p SolanaOptionParameter) Value() any {
	return p.value
}

var _ ContractEventParameter = SolanaOptionParameter{}
