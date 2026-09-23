package svm

import (
	"context"
	"fmt"

	solanago "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type SolanaGoClient struct {
	rpcClient *rpc.Client
}

func NewSolanaGoClient(url string) *SolanaGoClient {
	return &SolanaGoClient{
		rpcClient: rpc.New(url),
	}
}

func (c *SolanaGoClient) GetRecentBlockhash(ctx context.Context) (string, error) {
	resp, err := c.rpcClient.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return "", fmt.Errorf("error retrieving latest blockhash: %w", err)
	}

	return resp.Value.Blockhash.String(), nil
}

func (c *SolanaGoClient) GetRecentPrioritizationFees(ctx context.Context) ([]uint64, error) {
	results, err := c.rpcClient.GetRecentPrioritizationFees(ctx, solanago.PublicKeySlice{})
	if err != nil {
		return nil, fmt.Errorf("error retrieving recent prioritization fees: %w", err)
	}

	fees := make([]uint64, len(results))
	for i, r := range results {
		fees[i] = r.PrioritizationFee
	}
	return fees, nil
}

func (c *SolanaGoClient) SimulateTransaction(ctx context.Context, serializedTx string) (uint64, error) {
	tx, err := solanago.TransactionFromBase64(serializedTx)
	if err != nil {
		return 0, fmt.Errorf("error deserializing transaction for simulation: %w", err)
	}

	resp, err := c.rpcClient.SimulateTransactionWithOpts(ctx, tx, &rpc.SimulateTransactionOpts{})
	if err != nil {
		return 0, fmt.Errorf("error simulating transaction: %w", err)
	}
	if resp.Value.Err != nil {
		return 0, fmt.Errorf("simulation failed: %v", resp.Value.Err)
	}
	if resp.Value.UnitsConsumed == nil {
		return 0, fmt.Errorf("simulation returned no compute unit data")
	}

	return *resp.Value.UnitsConsumed, nil
}

func (c *SolanaGoClient) SendTransaction(ctx context.Context, signedTransaction string) (string, error) {
	tx, err := solanago.TransactionFromBase64(signedTransaction)
	if err != nil {
		return "", fmt.Errorf("error deserializing signed transaction: %w", err)
	}

	sig, err := c.rpcClient.SendTransactionWithOpts(ctx, tx, rpc.TransactionOpts{
		SkipPreflight: true,
	})
	if err != nil {
		return "", fmt.Errorf("error sending transaction to network: %w", err)
	}

	return sig.String(), nil
}

func (c *SolanaGoClient) Close() error {
	return c.rpcClient.Close()
}
