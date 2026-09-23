package nonce

import (
	"dlt-ingress/src/main/dltingress/domain/nonce"

	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/basemodel"
)

type NonceConfigurator func(*nonce.Nonce)

type NonceTestFactory struct{}

func NewNonceTestFactory() *NonceTestFactory {
	return &NonceTestFactory{}
}

func (f *NonceTestFactory) CreateEntity(config ...NonceConfigurator) *nonce.Nonce {
	value, _ := amount.NewFromString("0")

	entity := &nonce.Nonce{
		Model: basemodel.Model{
			Id: uuid.New(),
		},
		DltAccountId: "0xAbCdEf1234567890AbCdEf1234567890AbCdEf12",
		NetworkId:    "network-id-123",
		Value:        value,
	}
	for _, c := range config {
		c(entity)
	}
	return entity
}
