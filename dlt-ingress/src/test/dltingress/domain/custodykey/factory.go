package custodykey

import (
	"dlt-ingress/src/main/dltingress/domain/common"
	"dlt-ingress/src/main/dltingress/domain/custodykey"

	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/basemodel"
)

type CustodyKeyConfigurator func(*custodykey.CustodyKey)

type CustodyKeyTestFactory struct{}

func NewCustodyKeyTestFactory() *CustodyKeyTestFactory {
	return &CustodyKeyTestFactory{}
}

func (f *CustodyKeyTestFactory) CreateEntity(config ...CustodyKeyConfigurator) *custodykey.CustodyKey {
	key := &custodykey.CustodyKey{
		Model: basemodel.Model{
			Id: uuid.New(),
		},
		KeyType:         common.ECDSASecp256k1,
		Status:          custodykey.StatusActive,
		DltAccountId:    "0xAbCdEf1234567890AbCdEf1234567890AbCdEf12",
		Dlt:             common.EVM,
		ExternalId:      "ext-id-123",
		CustodyProvider: custodykey.CustodyProviderDFNS,
	}
	for _, c := range config {
		c(key)
	}
	return key
}
