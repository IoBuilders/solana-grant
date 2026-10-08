package nonce

import (
	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/base"
)

type NonceConfigurator func(*Nonce)

type NonceTestFactory struct{}

func NewNonceTestFactory() *NonceTestFactory {
	return &NonceTestFactory{}
}

func (f *NonceTestFactory) CreateEntity(config ...NonceConfigurator) *Nonce {
	value, _ := amount.NewFromString("0")

	entity := &Nonce{
		Entity: base.Entity{
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
