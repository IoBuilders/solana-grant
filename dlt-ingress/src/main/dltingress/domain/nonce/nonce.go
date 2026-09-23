package nonce

import (
	"dlt-ingress/src/main/dltingress/domain/domainerrors"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/basemodel"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/validate"
)

type Nonce struct {
	basemodel.Model
	DltAccountId string         `gorm:"type:varchar(255) not null;uniqueIndex:idx_nonce_dltaccount_network"`
	NetworkId    string         `gorm:"type:varchar(255) not null;uniqueIndex:idx_nonce_dltaccount_network"`
	Value        *amount.Amount `gorm:"type:varchar(100) not null"`
}

func NewNonce(
	dltAccountId string,
	networkId string,
	value *amount.Amount,
) (*Nonce, error) {
	if err := validate.StringMaxLength(dltAccountId, "DltAccountId", "Nonce", 255); err != nil {
		return nil, err
	}
	if err := validate.StringMaxLength(networkId, "NetworkId", "Nonce", 255); err != nil {
		return nil, err
	}
	if err := validateNonceValue(value); err != nil {
		return nil, err
	}

	return &Nonce{
		DltAccountId: dltAccountId,
		NetworkId:    networkId,
		Value:        value,
	}, nil
}

func (n *Nonce) UpdateValue(newValue *amount.Amount) error {
	if err := validateNonceValue(newValue); err != nil {
		return err
	}
	n.Value = newValue
	return nil
}

func validateNonceValue(value *amount.Amount) error {
	if value.LessThan(*amount.Zero()) || value.Decimals() != 0 {
		return domainerrors.NewInvalidNonceValueDomainError(value)
	}

	return nil
}
