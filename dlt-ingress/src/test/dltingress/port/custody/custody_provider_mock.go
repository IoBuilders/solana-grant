package custodymocks

import (
	"context"
	"dlt-ingress/src/main/dltingress/port/custody"

	"github.com/stretchr/testify/mock"
)

type CustodyProviderMock struct {
	mock.Mock
}

func (m *CustodyProviderMock) CreateKey(ctx context.Context, request *custody.CreateKeyRequest) (*custody.KeyResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*custody.KeyResponse), args.Error(1)
}

func (m *CustodyProviderMock) Sign(ctx context.Context, request *custody.SignRequest) (*custody.SignResponse, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*custody.SignResponse), args.Error(1)
}

var _ custody.Port = (*CustodyProviderMock)(nil)
