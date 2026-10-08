package evm

import (
	"context"
	"math/big"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/infra/noderatelimit"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/ratelimit"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
)

// JSON-RPC methods issued under the hood by the typed Client methods, used to pick their rate limit bucket.
const (
	methodGetTransactionCount = "eth_getTransactionCount"
	methodGetBlockByNumber    = "eth_getBlockByNumber"
	methodGasPrice            = "eth_gasPrice"
	methodGetBalance          = "eth_getBalance"
	methodNetVersion          = "net_version"
	methodCall                = "eth_call"
)

type rateLimitedClient struct {
	inner Client
	guard *noderatelimit.Guard
}

func NewRateLimitedClient(inner Client, limiter *ratelimit.RateLimiter, rateLimitConfig *config.RateLimitConfig) Client {
	return &rateLimitedClient{
		inner: inner,
		guard: noderatelimit.NewGuard(limiter, rateLimitConfig),
	}
}

func (c *rateLimitedClient) CallRpcMethod(ctx context.Context, result interface{}, method string, args ...interface{}) error {
	_, err := noderatelimit.Call(ctx, c.guard, method, func() (struct{}, error) {
		return struct{}{}, c.inner.CallRpcMethod(ctx, result, method, args...)
	})
	return err
}

func (c *rateLimitedClient) PendingNonceAt(ctx context.Context, account string) (uint64, error) {
	return noderatelimit.Call(ctx, c.guard, methodGetTransactionCount, func() (uint64, error) {
		return c.inner.PendingNonceAt(ctx, account)
	})
}

func (c *rateLimitedClient) GetBaseFeePerGas(ctx context.Context) (*amount.Amount, error) {
	return noderatelimit.Call(ctx, c.guard, methodGetBlockByNumber, func() (*amount.Amount, error) {
		return c.inner.GetBaseFeePerGas(ctx)
	})
}

func (c *rateLimitedClient) GetGasPrice(ctx context.Context) (*amount.Amount, error) {
	return noderatelimit.Call(ctx, c.guard, methodGasPrice, func() (*amount.Amount, error) {
		return c.inner.GetGasPrice(ctx)
	})
}

func (c *rateLimitedClient) GetBalance(ctx context.Context, account string) (*amount.Amount, error) {
	return noderatelimit.Call(ctx, c.guard, methodGetBalance, func() (*amount.Amount, error) {
		return c.inner.GetBalance(ctx, account)
	})
}

func (c *rateLimitedClient) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return noderatelimit.Call(ctx, c.guard, methodGetBlockByNumber, func() (*types.Header, error) {
		return c.inner.HeaderByNumber(ctx, number)
	})
}

func (c *rateLimitedClient) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	return noderatelimit.Call(ctx, c.guard, methodGetBlockByNumber, func() (*types.Block, error) {
		return c.inner.BlockByNumber(ctx, number)
	})
}

func (c *rateLimitedClient) NetworkID(ctx context.Context) (*big.Int, error) {
	return noderatelimit.Call(ctx, c.guard, methodNetVersion, func() (*big.Int, error) {
		return c.inner.NetworkID(ctx)
	})
}

func (c *rateLimitedClient) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	return noderatelimit.Call(ctx, c.guard, methodCall, func() ([]byte, error) {
		return c.inner.CallContract(ctx, msg, blockNumber)
	})
}

func (c *rateLimitedClient) Close() {
	c.inner.Close()
}
