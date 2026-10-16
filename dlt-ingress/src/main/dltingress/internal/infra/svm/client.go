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

// SignatureStatus is what a node reports about a transaction it has a record of.
type SignatureStatus struct {
	// Failed is true when the transaction was included in a block but its execution returned an error.
	Failed bool
	// ConfirmationStatus is the cluster's commitment for it: processed, confirmed or finalized.
	ConfirmationStatus string
}

type Client interface {
	GetRecentBlockhash(ctx context.Context) (string, error)
	GetRecentPrioritizationFees(ctx context.Context, accounts []string) ([]uint64, error)
	SimulateTransaction(ctx context.Context, serializedTx string) (uint64, error)
	SendTransaction(ctx context.Context, signedTransaction string) (string, error)
	GetSignatureStatus(ctx context.Context, signature string) (*SignatureStatus, error)
	IsBlockhashValid(ctx context.Context, blockhash string) (bool, error)
	GetBalance(ctx context.Context, account string) (*amount.Amount, error)
	GetHealth(ctx context.Context) error
	Close() error
}
