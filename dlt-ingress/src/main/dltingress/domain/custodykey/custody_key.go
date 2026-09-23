package custodykey

import (
	"dlt-ingress/src/main/dltingress/domain/common"
	"dlt-ingress/src/main/dltingress/domain/domainerrors"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/basemodel"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/validate"
)

type Status string

const (
	StatusActive       Status = "ACTIVE"
	StatusDiscontinued Status = "DISCONTINUED"
)

var validStatuses = map[Status]struct{}{
	StatusActive:       {},
	StatusDiscontinued: {},
}

func ParseStatus(value string) (Status, error) {
	s := Status(value)
	if _, ok := validStatuses[s]; !ok {
		return "", domainerrors.NewInvalidStatusDomainError(value)
	}
	return s, nil
}

type CustodyProvider string

const (
	CustodyProviderDFNS CustodyProvider = "DFNS"
	CustodyProviderKMS  CustodyProvider = "KMS"
)

var validCustodyProviders = map[CustodyProvider]struct{}{
	CustodyProviderDFNS: {},
	CustodyProviderKMS:  {},
}

func ParseCustodyProvider(value string) (CustodyProvider, error) {
	c := CustodyProvider(value)
	if _, ok := validCustodyProviders[c]; !ok {
		return "", domainerrors.NewInvalidCustodyProviderDomainError(value)
	}
	return c, nil
}

type CustodyKey struct {
	basemodel.Model
	KeyType         common.KeyType  `gorm:"type:varchar(30) not null"`
	Status          Status          `gorm:"type:varchar(20) not null;default:ACTIVE"`
	DltAccountId    string          `gorm:"type:varchar(255) not null;uniqueIndex"`
	Dlt             common.Dlt      `gorm:"type:varchar(20) not null"`
	ExternalId      string          `gorm:"type:varchar(255) not null"`
	CustodyProvider CustodyProvider `gorm:"type:varchar(30) not null"`
}

func NewCustodyKey(
	keyType string,
	dltAccountId string,
	dlt string,
	externalId string,
	custodyProvider string,
) (*CustodyKey, error) {
	parsedKeyType, err := common.ParseKeyType(keyType)
	if err != nil {
		return nil, err
	}
	parsedDlt, err := common.ParseDlt(dlt)
	if err != nil {
		return nil, err
	}
	parsedCustodyProvider, err := ParseCustodyProvider(custodyProvider)
	if err != nil {
		return nil, err
	}
	if err := validate.StringMaxLength(dltAccountId, "DltAccountId", "CustodyKey", 255); err != nil {
		return nil, err
	}
	if err := validate.StringMaxLength(externalId, "ExternalId", "CustodyKey", 255); err != nil {
		return nil, err
	}

	return &CustodyKey{
		KeyType:         parsedKeyType,
		Status:          StatusActive,
		DltAccountId:    dltAccountId,
		Dlt:             parsedDlt,
		ExternalId:      externalId,
		CustodyProvider: parsedCustodyProvider,
	}, nil
}

func (k *CustodyKey) Discontinue() {
	k.Status = StatusDiscontinued
}
