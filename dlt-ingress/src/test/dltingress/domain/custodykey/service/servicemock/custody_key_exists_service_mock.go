package servicemock

import (
	"context"
	"dlt-ingress/src/main/dltingress/domain/custodykey"

	"github.com/stretchr/testify/mock"
)

type CustodyKeyExistsMock struct {
	mock.Mock
}

func (m *CustodyKeyExistsMock) Execute(ctx context.Context, dltAccountId string) (*custodykey.CustodyKey, error) {
	args := m.Called(ctx, dltAccountId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*custodykey.CustodyKey), nil
}
