package mocks

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"

	"github.com/stretchr/testify/mock"
)

type EvmClientMock struct {
	mock.Mock
}

func (m *EvmClientMock) CallRpcMethod(ctx context.Context, result interface{}, method string, args ...interface{}) error {
	argsInt := m.Called(ctx, result, method, args)
	return argsInt.Error(0)
}

func (m *EvmClientMock) PendingNonceAt(ctx context.Context, account string) (uint64, error) {
	args := m.Called(ctx, account)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *EvmClientMock) GetBaseFeePerGas(ctx context.Context) (*amount.Amount, error) {
	args := m.Called(ctx)
	return args.Get(0).(*amount.Amount), args.Error(1)
}

func (m *EvmClientMock) GetGasPrice(ctx context.Context) (*amount.Amount, error) {
	args := m.Called(ctx)
	return args.Get(0).(*amount.Amount), args.Error(1)
}

func (m *EvmClientMock) Close() {
	m.Called()
}
