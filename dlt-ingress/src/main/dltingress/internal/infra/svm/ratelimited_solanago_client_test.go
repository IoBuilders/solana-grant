//go:build test

package svm_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/infra/svm"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/ratelimit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests run the real solana-go JSON-RPC transport through SolanaGoClient against a fake node
// answering 429. solana-go drops the headers of an error response and, when its body is a valid
// JSON-RPC error, the status too, so the 429 is only detected thanks to noderatelimit's transport.

func newFakeSolanaNode(t *testing.T, body string, headers map[string]string) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	requests := &atomic.Int32{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		for k, v := range headers {
			w.Header().Set(k, v)
		}
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return server, requests
}

func TestDltIngressSvmRateLimitedSolanaGoClient_NodeRateLimitBlocksTheBucket(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"json-rpc error body", `{"jsonrpc":"2.0","error":{"code":429,"message":"Too many requests for a specific RPC call"},"id":1}`},
		{"non json body", `<html>Too Many Requests</html>`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, requests := newFakeSolanaNode(t, tt.body, map[string]string{"Retry-After": "60"})
			solanaGoClient := svm.NewSolanaGoClient(node.URL)
			t.Cleanup(func() { _ = solanaGoClient.Close() })
			client := svm.NewRateLimitedClient(solanaGoClient, defaultBucket(100), &config.RateLimitConfig{
				RetryAfterHeader:       "Retry-After",
				RetryAfterHeaderFormat: config.RetryAfterHeaderFormatSeconds,
			})

			_, err := client.GetRecentBlockhash(context.Background())
			exceededErr, ok := errors.AsType[*ratelimit.ExceededError](err)
			require.True(t, ok, "expected the node's 429 to surface as ExceededError, got %v", err)
			assert.InDelta(t, 60*time.Second, exceededErr.RetryAfter, float64(time.Second), "retry after comes from the node's Retry-After")

			_, err = client.GetRecentBlockhash(context.Background())
			assert.True(t, ratelimit.IsExceededError(err), "Retry-After must block the bucket")
			assert.Equal(t, int32(1), requests.Load(), "the blocked call must not reach the node")
		})
	}
}
