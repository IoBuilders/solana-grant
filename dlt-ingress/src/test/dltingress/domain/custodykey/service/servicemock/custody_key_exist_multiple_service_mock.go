package servicemock

import (
	"context"
	"dlt-ingress/src/main/dltingress/domain/custodykey"

	"github.com/stretchr/testify/mock"
)

type CustodyKeyExistMultipleMock struct {
	mock.Mock
}

func (m *CustodyKeyExistMultipleMock) Execute(ctx context.Context, dltAccountIds []string) ([]*custodykey.CustodyKey, error) {
	args := m.Called(ctx, dltAccountIds)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*custodykey.CustodyKey), args.Error(1)
}
