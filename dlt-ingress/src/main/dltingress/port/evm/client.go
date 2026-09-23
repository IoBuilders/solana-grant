package evm

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
)

type Client interface {
	CallRpcMethod(ctx context.Context, result interface{}, method string, args ...interface{}) error
	PendingNonceAt(ctx context.Context, account string) (uint64, error)
	GetBaseFeePerGas(ctx context.Context) (*amount.Amount, error)
	GetGasPrice(ctx context.Context) (*amount.Amount, error)
	Close()
}
