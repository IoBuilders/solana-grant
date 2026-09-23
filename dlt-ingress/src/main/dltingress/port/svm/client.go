package svm

import "context"

type Client interface {
	GetRecentBlockhash(ctx context.Context) (string, error)
	GetRecentPrioritizationFees(ctx context.Context) ([]uint64, error)
	SimulateTransaction(ctx context.Context, serializedTx string) (uint64, error)
	SendTransaction(ctx context.Context, signedTransaction string) (string, error)
	Close() error
}
