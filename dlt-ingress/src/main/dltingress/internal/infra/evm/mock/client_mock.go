package mock

import (
	"context"
	"math/big"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
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

func (m *EvmClientMock) GetBalance(ctx context.Context, account string) (*amount.Amount, error) {
	args := m.Called(ctx, account)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*amount.Amount), args.Error(1)
}

func (m *EvmClientMock) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	args := m.Called(ctx, number)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Header), args.Error(1)
}

func (m *EvmClientMock) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	args := m.Called(ctx, number)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Block), args.Error(1)
}

func (m *EvmClientMock) NetworkID(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *EvmClientMock) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	args := m.Called(ctx, msg, blockNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *EvmClientMock) Close() {
	m.Called()
}
