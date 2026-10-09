package mock

import (
	"context"

	"dlt-ingress/src/main/dltingress/internal/infra/evm"
	"github.com/stretchr/testify/mock"
)

type EvmClientRegistryMock struct {
	mock.Mock
}

func (m *EvmClientRegistryMock) GetClientForNetworkId(ctx context.Context, networkId string) (evm.Client, error) {
	args := m.Called(ctx, networkId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(evm.Client), args.Error(1)
}

func (m *EvmClientRegistryMock) Shutdown() {
	m.Called()
}
