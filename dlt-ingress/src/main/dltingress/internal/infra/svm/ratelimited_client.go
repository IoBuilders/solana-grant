package svm

import (
	"context"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/infra/noderatelimit"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/ratelimit"
)

// JSON-RPC methods issued under the hood by the typed Client methods, used to pick their rate limit bucket.
const (
	methodGetLatestBlockhash          = "getLatestBlockhash"
	methodGetRecentPrioritizationFees = "getRecentPrioritizationFees"
	methodSimulateTransaction         = "simulateTransaction"
	methodSendTransaction             = "sendTransaction"
	methodGetBalance                  = "getBalance"
	methodGetHealth                   = "getHealth"
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

func (c *rateLimitedClient) GetRecentBlockhash(ctx context.Context) (string, error) {
	return noderatelimit.Call(ctx, c.guard, methodGetLatestBlockhash, func() (string, error) {
		return c.inner.GetRecentBlockhash(ctx)
	})
}

func (c *rateLimitedClient) GetRecentPrioritizationFees(ctx context.Context, accounts []string) ([]uint64, error) {
	return noderatelimit.Call(ctx, c.guard, methodGetRecentPrioritizationFees, func() ([]uint64, error) {
		return c.inner.GetRecentPrioritizationFees(ctx, accounts)
	})
}

func (c *rateLimitedClient) SimulateTransaction(ctx context.Context, serializedTx string) (uint64, error) {
	return noderatelimit.Call(ctx, c.guard, methodSimulateTransaction, func() (uint64, error) {
		return c.inner.SimulateTransaction(ctx, serializedTx)
	})
}

func (c *rateLimitedClient) SendTransaction(ctx context.Context, signedTransaction string) (string, error) {
	return noderatelimit.Call(ctx, c.guard, methodSendTransaction, func() (string, error) {
		return c.inner.SendTransaction(ctx, signedTransaction)
	})
}

func (c *rateLimitedClient) GetBalance(ctx context.Context, account string) (*amount.Amount, error) {
	return noderatelimit.Call(ctx, c.guard, methodGetBalance, func() (*amount.Amount, error) {
		return c.inner.GetBalance(ctx, account)
	})
}

func (c *rateLimitedClient) GetHealth(ctx context.Context) error {
	_, err := noderatelimit.Call(ctx, c.guard, methodGetHealth, func() (struct{}, error) {
		return struct{}{}, c.inner.GetHealth(ctx)
	})
	return err
}

func (c *rateLimitedClient) Close() error {
	return c.inner.Close()
}
