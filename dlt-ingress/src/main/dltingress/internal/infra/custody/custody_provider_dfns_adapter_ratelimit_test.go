//go:build test

package custody

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"dlt-ingress/src/main/dltingress/internal/infra/noderatelimit"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/ratelimit"

	"github.com/dfns/dfns-sdk-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func defaultBucketGuard(requestsPerMinute int, retryAfterHeader string) *noderatelimit.Guard {
	limiter := ratelimit.New(ratelimit.WithBucket("DEFAULT", requestsPerMinute, ratelimit.MatchAll()))
	return noderatelimit.NewGuard(limiter, &config.RateLimitConfig{
		RetryAfterHeader:       retryAfterHeader,
		RetryAfterHeaderFormat: config.RetryAfterHeaderFormatSeconds,
	})
}

// newFakeDfns starts a TLS server (the SDK only accepts https) whose handler also gets the request number,
// and a DFNS client wired like NewDfnsClient.
func newFakeDfns(t *testing.T, handler func(w http.ResponseWriter, r *http.Request, n int32)) (*dfns.Client, *atomic.Int32) {
	t.Helper()
	requests := &atomic.Int32{}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, requests.Add(1))
	}))
	t.Cleanup(server.Close)
	client, err := dfns.NewClient(dfns.Options{
		BaseURL:    server.URL,
		AuthToken:  "token",
		HTTPClient: &http.Client{Transport: noderatelimit.NewTransport(server.Client().Transport)},
	})
	require.NoError(t, err)
	return client, requests
}

func setSignPolling(t *testing.T, timeout, interval time.Duration) {
	t.Helper()
	previousConfig := config.DltIngressConfig
	t.Cleanup(func() { config.DltIngressConfig = previousConfig })
	config.DltIngressConfig = &config.Config{}
	config.DltIngressConfig.DltIngress.Custody.SignPollingTimeout = timeout
	config.DltIngressConfig.DltIngress.Custody.SignPollingInterval = interval
}

func writeSignature(w http.ResponseWriter, status string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"id":"sig-1","keyId":"key-1","status":"` + status + `","signature":{"encoded":"0xabcd"}}`))
}

func TestDltIngressDfnsProvider_RejectsWithoutCallingDfnsWhenBucketExhausted(t *testing.T) {
	client, requests := newFakeDfns(t, func(http.ResponseWriter, *http.Request, int32) {})
	limiter := ratelimit.New(ratelimit.WithBucket("WALLETS", 1, ratelimit.Methods(dfnsMethodCreateWallet)))
	allowed, _ := limiter.Allow(context.Background(), dfnsMethodCreateWallet)
	require.True(t, allowed)
	provider := NewDfnsProvider(client, noderatelimit.NewGuard(limiter, nil))

	_, err := provider.CreateKey(context.Background(), &CreateKeyRequest{Dlt: string(common.EVM)})

	exceededErr, ok := errors.AsType[*ratelimit.ExceededError](err)
	require.True(t, ok, "expected ExceededError, got %v", err)
	assert.Equal(t, dfnsMethodCreateWallet, exceededErr.Method)
	assert.Equal(t, int32(0), requests.Load(), "the rejected call must not reach DFNS")
}

func TestDltIngressDfnsProvider_SignaturePollWaitsOutTheRateLimitInsteadOfFailing(t *testing.T) {
	setSignPolling(t, 5*time.Second, 50*time.Millisecond)
	client, requests := newFakeDfns(t, func(w http.ResponseWriter, _ *http.Request, n int32) {
		if n == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		writeSignature(w, "Signed")
	})

	signature, err := NewDfnsProvider(client, defaultBucketGuard(100, "Retry-After")).pendingSignaturePoll(context.Background(), "key-1", "sig-1")

	require.NoError(t, err)
	assert.Equal(t, "0xabcd", signature["encoded"])
	assert.Equal(t, int32(2), requests.Load(), "the ticks within Retry-After must not reach DFNS")
}

func TestDltIngressDfnsProvider_SignaturePollDoesNotCallDfnsWhileBucketExhausted(t *testing.T) {
	setSignPolling(t, 300*time.Millisecond, 50*time.Millisecond)
	client, requests := newFakeDfns(t, func(w http.ResponseWriter, _ *http.Request, _ int32) {
		writeSignature(w, "Pending")
	})

	_, err := NewDfnsProvider(client, defaultBucketGuard(1, "")).pendingSignaturePoll(context.Background(), "key-1", "sig-1")

	assert.ErrorIs(t, err, context.DeadlineExceeded, "the poll keeps waiting until its own timeout")
	assert.Equal(t, int32(1), requests.Load(), "only the first tick fits the bucket; the rest are rejected locally")
}
