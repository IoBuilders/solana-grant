package svm

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"

	"github.com/stretchr/testify/mock"
)

type SvmClientMock struct {
	mock.Mock
}

func (m *SvmClientMock) GetRecentBlockhash(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *SvmClientMock) GetRecentPrioritizationFees(ctx context.Context, accounts []string) ([]uint64, error) {
	args := m.Called(ctx, accounts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uint64), args.Error(1)
}

func (m *SvmClientMock) SimulateTransaction(ctx context.Context, serializedTx string) (uint64, error) {
	args := m.Called(ctx, serializedTx)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *SvmClientMock) SendTransaction(ctx context.Context, signedTransaction string) (string, error) {
	args := m.Called(ctx, signedTransaction)
	return args.String(0), args.Error(1)
}

func (m *SvmClientMock) GetBalance(ctx context.Context, account string) (*amount.Amount, error) {
	args := m.Called(ctx, account)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*amount.Amount), args.Error(1)
}

func (m *SvmClientMock) GetHealth(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *SvmClientMock) Close() error {
	args := m.Called()
	return args.Error(0)
}
