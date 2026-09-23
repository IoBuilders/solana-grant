package solana

import "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"

// PublicKeyParameterDefinition declares an account-address-typed field
// (Anchor's pubkey/publicKey IDL type) — a fixed 32-byte value, base58
// encoded on decode, no further decode metadata needed.
type PublicKeyParameterDefinition struct {
	parameterDefinitionBase
}

func NewPublicKeyParameterDefinition(position int) (PublicKeyParameterDefinition, error) {
	base, err := newParameterDefinitionBase(position)
	if err != nil {
		return PublicKeyParameterDefinition{}, err
	}
	return PublicKeyParameterDefinition{parameterDefinitionBase: base}, nil
}

func (d PublicKeyParameterDefinition) Type() parameter.Type {
	return parameter.TypePublicKey
}

var _ ParameterDefinition = PublicKeyParameterDefinition{}
