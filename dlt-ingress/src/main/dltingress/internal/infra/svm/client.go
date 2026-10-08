package svm

import (
	"context"
	"errors"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
)

// ErrBlockhashNotFound marks a SimulateTransaction failure caused specifically by a stale recent
// blockhash the node no longer recognizes, so callers can tell it apart from any other simulation
// failure and choose to refetch a fresh blockhash and retry once, instead of giving up.
var ErrBlockhashNotFound = errors.New("blockhash not found")

type Client interface {
	GetRecentBlockhash(ctx context.Context) (string, error)
	GetRecentPrioritizationFees(ctx context.Context, accounts []string) ([]uint64, error)
	SimulateTransaction(ctx context.Context, serializedTx string) (uint64, error)
	SendTransaction(ctx context.Context, signedTransaction string) (string, error)
	GetBalance(ctx context.Context, account string) (*amount.Amount, error)
	GetHealth(ctx context.Context) error
	Close() error
}
