package common

import (
	"dlt-ingress/src/main/dltingress/domain/domainerrors"
)

type Dlt string

const (
	EVM       Dlt = "EVM"
	Hashgraph Dlt = "Hashgraph"
	SVM       Dlt = "SVM"
)

var validDlts = map[Dlt]struct{}{
	EVM:       {},
	Hashgraph: {},
	SVM:       {},
}

func ParseDlt(value string) (Dlt, error) {
	d := Dlt(value)
	if _, ok := validDlts[d]; !ok {
		return "", domainerrors.NewInvalidDltDomainError(value)
	}
	return d, nil
}

var dltKeyTypes = map[Dlt]KeyType{
	EVM:       ECDSASecp256k1,
	Hashgraph: ECDSASecp256k1,
	SVM:       ED25519,
}

func (d Dlt) KeyType() KeyType {
	return dltKeyTypes[d]
}

type KeyType string

const (
	ECDSASecp256k1 KeyType = "ECDSA_SECP256K1"
	ED25519        KeyType = "ED25519"
)

var validKeyTypes = map[KeyType]struct{}{
	ECDSASecp256k1: {},
	ED25519:        {},
}

func ParseKeyType(value string) (KeyType, error) {
	k := KeyType(value)
	if _, ok := validKeyTypes[k]; !ok {
		return "", domainerrors.NewInvalidKeyTypeDomainError(value)
	}
	return k, nil
}
