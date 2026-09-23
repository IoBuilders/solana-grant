package solana

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/interactor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/block"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/common"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/httpclient"
)

// SolanaRpcBlockInteractor implements interactor.BlockInteractor against a
// Solana node's JSON-RPC HTTP endpoint. Each GetBlock/GetSlot call is a
// single, independent HTTP request: a failure is returned immediately rather
// than retried internally, since callers (e.g. SlotProcessorPermanentTrigger)
// already apply their own retry policy around these methods.
type SolanaRpcBlockInteractor struct {
	endpoint   string
	httpClient *http.Client
}

// NewSolanaRpcBlockInteractor builds a SolanaRpcBlockInteractor targeting
// connection's endpoint. connection must be an HTTP connection (a Node may
// also be configured over WS for other purposes, e.g. slot subscriptions,
// but getBlock/getSlot are plain HTTP JSON-RPC calls here). client
// configures the underlying HTTP transport's connection pooling and
// timeouts.
func NewSolanaRpcBlockInteractor(connection common.Connection, client *httpclient.HttpClient) (*SolanaRpcBlockInteractor, error) {
	if connection == nil {
		return nil, domainerrors.NewEmptyFieldError("Connection", "SolanaRpcBlockInteractor")
	}
	if err := connection.Validate(); err != nil {
		return nil, err
	}
	if connection.ConnectionType() != common.ConnectionTypeHttp {
		return nil, domainerrors.NewInvalidFieldError(
			"Connection", "SolanaRpcBlockInteractor",
			fmt.Sprintf("must be an HTTP connection, got %s", connection.ConnectionType()),
		)
	}
	if client == nil {
		return nil, domainerrors.NewEmptyFieldError("Client", "SolanaRpcBlockInteractor")
	}
	return &SolanaRpcBlockInteractor{
		endpoint:   connection.ConnectionEndpoint().URL(),
		httpClient: newHTTPClient(client),
	}, nil
}

func newHTTPClient(cfg *httpclient.HttpClient) *http.Client {
	dialer := &net.Dialer{Timeout: cfg.ConnectTimeout}
	transport := &http.Transport{
		DialContext:         dialer.DialContext,
		MaxIdleConnsPerHost: cfg.MaxIdleConnections,
		IdleConnTimeout:     cfg.KeepAliveDuration,
	}
	return &http.Client{Transport: transport, Timeout: cfg.CallTimeout}
}

// Type returns interactor.TypeBlock.
func (i *SolanaRpcBlockInteractor) Type() interactor.Type {
	return interactor.TypeBlock
}

// GetBlock calls getBlock for slot, requesting jsonParsed encoding and no
// rewards, with maxSupportedTransactionVersion set so blocks containing
// versioned transactions don't fail the whole call. A null RPC result (the
// leader skipped slot) returns (nil, nil); a JSON-RPC error result is mapped
// to a typed error by mapRPCError.
func (i *SolanaRpcBlockInteractor) GetBlock(ctx context.Context, slot uint64) (*block.SolanaBlock, error) {
	params := []any{
		slot,
		map[string]any{
			"encoding":                       "jsonParsed",
			"rewards":                        false,
			"maxSupportedTransactionVersion": 0,
			"commitment":                     "confirmed",
		},
	}
	var raw *rpcBlock
	if err := i.call(ctx, "getBlock", params, &raw); err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, nil
	}
	return raw.toDomain(slot)
}

// GetSlot calls getSlot at confirmed commitment, the same commitment level
// GetBlock implicitly resolves against.
func (i *SolanaRpcBlockInteractor) GetSlot(ctx context.Context) (uint64, error) {
	params := []any{map[string]any{"commitment": "confirmed"}}
	var slot uint64
	if err := i.call(ctx, "getSlot", params, &slot); err != nil {
		return 0, err
	}
	return slot, nil
}

var _ interactor.BlockInteractor = (*SolanaRpcBlockInteractor)(nil)
