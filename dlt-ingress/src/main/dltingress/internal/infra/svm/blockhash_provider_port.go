package svm

import "context"

type BlockhashProvider interface {
	GetRecentBlockhash(ctx context.Context, networkId string) (string, error)

	// InvalidateRecentBlockhash discards any cached blockhash for networkId, so the next
	// GetRecentBlockhash call is forced to fetch a fresh one. Callers use this after a node rejects
	// a transaction for carrying a blockhash it no longer recognizes.
	InvalidateRecentBlockhash(ctx context.Context, networkId string) error
}
