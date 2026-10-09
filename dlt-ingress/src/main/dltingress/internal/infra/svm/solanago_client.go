package svm

import (
	"context"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"time"

	"dlt-ingress/src/main/dltingress/internal/infra/noderatelimit"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
	"github.com/klauspost/compress/gzhttp"
)

// Same values solana-go uses for its default HTTP client (rpc.New), which it does not export.
const (
	httpTimeout             = 5 * time.Minute
	httpKeepAlive           = 180 * time.Second
	httpMaxConnsPerHost     = 9
	httpTLSHandshakeTimeout = 10 * time.Second
)

// blockhashNotFoundReason is the literal value Solana nodes put in a simulation's top-level
// TransactionError when the submitted recent blockhash is no longer known to the cluster.
const blockhashNotFoundReason = "BlockhashNotFound"

type SolanaGoClient struct {
	rpcClient *rpc.Client
}

// NewSolanaGoClient dials the node through noderatelimit's transport, so a 429 reaches the caller as a
// noderatelimit.TooManyRequestsError with its headers, which solana-go would otherwise drop.
func NewSolanaGoClient(url string) *SolanaGoClient {
	httpClient := &http.Client{
		Timeout:   httpTimeout,
		Transport: noderatelimit.NewTransport(gzhttp.Transport(newHTTPTransport())),
	}
	return &SolanaGoClient{
		rpcClient: rpc.NewWithCustomRPCClient(jsonrpc.NewClientWithOpts(url, &jsonrpc.RPCClientOpts{HTTPClient: httpClient})),
	}
}

func newHTTPTransport() *http.Transport {
	return &http.Transport{
		IdleConnTimeout:     httpTimeout,
		MaxConnsPerHost:     httpMaxConnsPerHost,
		MaxIdleConnsPerHost: httpMaxConnsPerHost,
		Proxy:               http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   httpTimeout,
			KeepAlive: httpKeepAlive,
		}).DialContext,
		ForceAttemptHTTP2:   true,
		TLSHandshakeTimeout: httpTLSHandshakeTimeout,
	}
}

func (c *SolanaGoClient) GetRecentBlockhash(ctx context.Context) (string, error) {
	resp, err := c.rpcClient.GetLatestBlockhash(ctx, rpc.CommitmentConfirmed)
	if err != nil {
		return "", fmt.Errorf("error retrieving latest blockhash: %w", err)
	}

	return resp.Value.Blockhash.String(), nil
}

func (c *SolanaGoClient) GetRecentPrioritizationFees(ctx context.Context, accounts []string) ([]uint64, error) {
	publicKeys := make(solana.PublicKeySlice, len(accounts))
	for i, account := range accounts {
		publicKey, err := solana.PublicKeyFromBase58(account)
		if err != nil {
			return nil, fmt.Errorf("error parsing account %s: %w", account, err)
		}
		publicKeys[i] = publicKey
	}

	results, err := c.rpcClient.GetRecentPrioritizationFees(ctx, publicKeys)
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
	tx, err := solana.TransactionFromBase64(serializedTx)
	if err != nil {
		return 0, fmt.Errorf("error deserializing transaction for simulation: %w", err)
	}

	resp, err := c.rpcClient.SimulateTransactionWithOpts(ctx, tx, &rpc.SimulateTransactionOpts{})
	if err != nil {
		return 0, fmt.Errorf("error simulating transaction: %w", err)
	}
	if resp.Value.Err != nil {
		if reason, ok := resp.Value.Err.(string); ok && reason == blockhashNotFoundReason {
			return 0, fmt.Errorf("%w: %v", ErrBlockhashNotFound, resp.Value.Err)
		}
		return 0, fmt.Errorf("simulation failed: %v", resp.Value.Err)
	}
	if resp.Value.UnitsConsumed == nil {
		return 0, fmt.Errorf("simulation returned no compute unit data")
	}

	return *resp.Value.UnitsConsumed, nil
}

func (c *SolanaGoClient) SendTransaction(ctx context.Context, signedTransaction string) (string, error) {
	tx, err := solana.TransactionFromBase64(signedTransaction)
	if err != nil {
		return "", fmt.Errorf("error deserializing signed transaction: %w", err)
	}

	sig, err := c.rpcClient.SendTransaction(ctx, tx)
	if err != nil {
		return "", fmt.Errorf("error sending transaction to network: %w", err)
	}

	return sig.String(), nil
}

func (c *SolanaGoClient) GetBalance(ctx context.Context, account string) (*amount.Amount, error) {
	publicKey, err := solana.PublicKeyFromBase58(account)
	if err != nil {
		return nil, fmt.Errorf("error parsing account %s: %w", account, err)
	}

	resp, err := c.rpcClient.GetBalance(ctx, publicKey, rpc.CommitmentFinalized)
	if err != nil {
		return nil, fmt.Errorf("error retrieving balance for account %s: %w", account, err)
	}

	result, err := amount.New(new(big.Int).SetUint64(resp.Value), 0)
	if err != nil {
		return nil, fmt.Errorf("invalid balance amount: %w", err)
	}
	return result, nil
}

func (c *SolanaGoClient) GetHealth(ctx context.Context) error {
	status, err := c.rpcClient.GetHealth(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving node health: %w", err)
	}
	if status != rpc.HealthOk {
		return fmt.Errorf("node reported health %q", status)
	}
	return nil
}

func (c *SolanaGoClient) Close() error {
	return c.rpcClient.Close()
}
