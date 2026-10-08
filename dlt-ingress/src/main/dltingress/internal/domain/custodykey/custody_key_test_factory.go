package custodykey

import (
	"dlt-ingress/src/main/dltingress/internal/domain/common"

	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/base"
)

type CustodyKeyConfigurator func(*CustodyKey)

type CustodyKeyTestFactory struct{}

func NewCustodyKeyTestFactory() *CustodyKeyTestFactory {
	return &CustodyKeyTestFactory{}
}

func (f *CustodyKeyTestFactory) CreateEntity(config ...CustodyKeyConfigurator) *CustodyKey {
	key := &CustodyKey{
		Entity: base.Entity{
			Id: uuid.New(),
		},
		KeyType:         common.ECDSASecp256k1,
		Status:          StatusActive,
		DltAccountId:    "0xAbCdEf1234567890AbCdEf1234567890AbCdEf12",
		Dlt:             common.EVM,
		ExternalId:      "ext-id-123",
		CustodyProvider: CustodyProviderDFNS,
	}
	for _, c := range config {
		c(key)
	}
	return key
}
