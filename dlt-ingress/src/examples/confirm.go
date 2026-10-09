package examples

import (
	"context"
	"fmt"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// WaitForConfirmation polls the signature's status until it reaches at least
// the "confirmed" commitment level. It returns an error if the transaction
// landed with an on-chain error, or if the timeout elapses before it confirms.
func WaitForConfirmation(ctx context.Context, rpcClient *rpc.Client, signature string, timeout time.Duration) error {
	sig, err := solana.SignatureFromBase58(signature)
	if err != nil {
		return fmt.Errorf("invalid signature %q: %w", signature, err)
	}

	deadline := time.Now().Add(timeout)
	for {
		statuses, err := rpcClient.GetSignatureStatuses(ctx, true, sig)
		if err != nil {
			return fmt.Errorf("get signature statuses for %s: %w", signature, err)
		}

		if status := statuses.Value[0]; status != nil {
			if status.Err != nil {
				return fmt.Errorf("transaction %s failed on-chain: %v", signature, status.Err)
			}
			if status.ConfirmationStatus == rpc.ConfirmationStatusConfirmed || status.ConfirmationStatus == rpc.ConfirmationStatusFinalized {
				return nil
			}
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("timed out after %s waiting for %s to confirm", timeout, signature)
		}
		time.Sleep(500 * time.Millisecond)
	}
}
