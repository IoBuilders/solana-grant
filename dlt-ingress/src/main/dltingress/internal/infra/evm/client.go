package evm

import (
	"context"
	"math/big"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
)

type Client interface {
	CallRpcMethod(ctx context.Context, result interface{}, method string, args ...interface{}) error
	PendingNonceAt(ctx context.Context, account string) (uint64, error)
	GetBaseFeePerGas(ctx context.Context) (*amount.Amount, error)
	GetGasPrice(ctx context.Context) (*amount.Amount, error)
	GetBalance(ctx context.Context, account string) (*amount.Amount, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
	NetworkID(ctx context.Context) (*big.Int, error)
	CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)
	Close()
}
