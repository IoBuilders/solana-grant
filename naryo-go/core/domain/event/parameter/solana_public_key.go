package parameter

// SolanaPublicKeyParameter is the Solana implementation of an account
// address argument (Anchor's pubkey/publicKey IDL type), the Solana
// analogue of an Ethereum address.
type SolanaPublicKeyParameter struct {
	solanaParameterBase
	value string
}

func NewSolanaPublicKeyParameter(position int, value string) (SolanaPublicKeyParameter, error) {
	base, err := newSolanaParameterBase(position)
	if err != nil {
		return SolanaPublicKeyParameter{}, err
	}
	return SolanaPublicKeyParameter{solanaParameterBase: base, value: value}, nil
}

func (p SolanaPublicKeyParameter) Type() Type {
	return TypePublicKey
}

func (p SolanaPublicKeyParameter) Value() any {
	return p.value
}

var _ ContractEventParameter = SolanaPublicKeyParameter{}
