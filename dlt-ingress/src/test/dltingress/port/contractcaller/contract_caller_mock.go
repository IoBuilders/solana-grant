//go:build test

package contractcaller

import (
	"context"
	"dlt-ingress/src/main/dltingress/port/contractcaller"

	"github.com/stretchr/testify/mock"
)

type ContractCallerMock struct {
	mock.Mock
}

func (m *ContractCallerMock) Call(ctx context.Context, request contractcaller.CallRequest) (*contractcaller.CallResponse, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contractcaller.CallResponse), args.Error(1)
}

var _ contractcaller.Port = (*ContractCallerMock)(nil)
