package nonceprovider

import (
	"context"
	"dlt-ingress/src/main/dltingress/port/nonceprovider"

	"github.com/stretchr/testify/mock"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
)

type NonceProviderMock struct {
	mock.Mock
}

func (m *NonceProviderMock) GetNonce(ctx context.Context, request *nonceprovider.GetNonceRequest) (*amount.Amount, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*amount.Amount), args.Error(1)
}

func (m *NonceProviderMock) SetNonce(ctx context.Context, request *nonceprovider.SetNonceRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}
