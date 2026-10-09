//go:build test

package evm_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/infra/evm"
	"dlt-ingress/src/main/dltingress/internal/infra/noderatelimit"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/ratelimit"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests run the real go-ethereum JSON-RPC transport through GoEthClient against a fake node
// answering 429, to pin down what a rate-limited response actually looks like from our side.

// rateLimitedBody is the body the Hedera JSON-RPC relay (v0.78.5) returns along with its HTTP 429.
const rateLimitedBody = `{"error":{"code":-32605,"message":"[Request ID: 884f8777-118e-4ac0-8617-90beae7a5f2b] IP Rate limit exceeded on eth_call"},"jsonrpc":"2.0","id":1}`

func newFakeNode(t *testing.T, status int, headers map[string]string) (*httptest.Server, *atomic.Int32) {
	requests := &atomic.Int32{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		for k, v := range headers {
			w.Header().Set(k, v)
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(rateLimitedBody))
	}))
	t.Cleanup(server.Close)
	return server, requests
}

func newRateLimitedGoEthClient(t *testing.T, url string, bucket ratelimit.OptionConfig, retryAfterHeader string, retryAfterHeaderFormat config.RetryAfterHeaderFormat) evm.Client {
	httpClient := &http.Client{Transport: noderatelimit.NewTransport(http.DefaultTransport)}
	goEthClient, err := evm.NewGoEthClient(context.Background(), url, httpClient)
	require.NoError(t, err)
	t.Cleanup(goEthClient.Close)
	return evm.NewRateLimitedClient(goEthClient, ratelimit.New(bucket), &config.RateLimitConfig{
		RetryAfterHeader:       retryAfterHeader,
		RetryAfterHeaderFormat: retryAfterHeaderFormat,
	})
}

func tier1(requestsPerMinute int) ratelimit.OptionConfig {
	return ratelimit.WithBucket("TIER_1", requestsPerMinute, ratelimit.Methods("eth_call"))
}

// callTwice makes a first call that the node answers, then a second one, and returns the second one's error.
func callTwice(t *testing.T, client evm.Client) error {
	var result string
	require.Error(t, client.CallRpcMethod(context.Background(), &result, "eth_call"))
	return client.CallRpcMethod(context.Background(), &result, "eth_call")
}

func requireExceeded(t *testing.T, err error) *ratelimit.ExceededError {
	var exceeded *ratelimit.ExceededError
	require.True(t, errors.As(err, &exceeded), "expected ratelimit.ExceededError, got %v", err)
	return exceeded
}

func TestDltIngressRateLimitedGoEthClient_Live429_ReturnsExceededError(t *testing.T) {
	server, requests := newFakeNode(t, http.StatusTooManyRequests, map[string]string{"Retry-After": "30"})
	client := newRateLimitedGoEthClient(t, server.URL, tier1(100), "Retry-After", config.RetryAfterHeaderFormatSeconds)

	var result string
	err := client.CallRpcMethod(context.Background(), &result, "eth_call")

	assert.Equal(t, 30, requireExceeded(t, err).RetryAfterSeconds(), "the node's 429 must come back as ratelimit.ExceededError")

	err = client.CallRpcMethod(context.Background(), &result, "eth_call")

	assert.True(t, ratelimit.IsExceededError(err), "expected the bucket to reject locally after the node's 429, got %v", err)
	assert.Equal(t, int32(1), requests.Load(), "the second call must not reach the node")
}

func TestDltIngressRateLimitedGoEthClient_RetryAfterSeconds_BlocksUntilDeclaredReset(t *testing.T) {
	server, requests := newFakeNode(t, http.StatusTooManyRequests, map[string]string{"Retry-After": "30"})
	client := newRateLimitedGoEthClient(t, server.URL, tier1(100), "Retry-After", config.RetryAfterHeaderFormatSeconds)

	exceeded := requireExceeded(t, callTwice(t, client))

	assert.InDelta(t, 30*time.Second, exceeded.RetryAfter, float64(time.Second))
	assert.Equal(t, int32(1), requests.Load())
}

func TestDltIngressRateLimitedGoEthClient_RetryAfterUnixTimestamp_BlocksUntilDeclaredReset(t *testing.T) {
	resetAt := time.Now().Add(45 * time.Second).Unix()
	server, _ := newFakeNode(t, http.StatusTooManyRequests, map[string]string{"X-RateLimit-Reset": strconv.FormatInt(resetAt, 10)})
	client := newRateLimitedGoEthClient(t, server.URL, tier1(100), "X-RateLimit-Reset", config.RetryAfterHeaderFormatUnixTimestamp)

	exceeded := requireExceeded(t, callTwice(t, client))

	assert.InDelta(t, 45*time.Second, exceeded.RetryAfter, float64(2*time.Second))
}

func TestDltIngressRateLimitedGoEthClient_RetryAfterThroughTypedMethod_BlocksItsBucket(t *testing.T) {
	server, requests := newFakeNode(t, http.StatusTooManyRequests, map[string]string{"Retry-After": "20"})
	client := newRateLimitedGoEthClient(t, server.URL, ratelimit.WithBucket("TIER_2", 800, ratelimit.Methods("eth_gasPrice")), "Retry-After", config.RetryAfterHeaderFormatSeconds)

	_, err := client.GetGasPrice(context.Background())
	require.Error(t, err)
	_, err = client.GetGasPrice(context.Background())

	exceeded := requireExceeded(t, err)
	assert.InDelta(t, 20*time.Second, exceeded.RetryAfter, float64(time.Second))
	assert.Equal(t, int32(1), requests.Load())
}

func TestDltIngressRateLimitedGoEthClient_RetryAfterNotConfigured_RefillsAtBucketRate(t *testing.T) {
	server, requests := newFakeNode(t, http.StatusTooManyRequests, map[string]string{"Retry-After": "30"})
	client := newRateLimitedGoEthClient(t, server.URL, tier1(100), "", "")

	exceeded := requireExceeded(t, callTwice(t, client))

	assert.Less(t, exceeded.RetryAfter, 5*time.Second, "without a configured header the 30s one is ignored")
	assert.Equal(t, int32(1), requests.Load())
}

func TestDltIngressRateLimitedGoEthClient_RetryAfterUnparseable_RefillsAtBucketRate(t *testing.T) {
	server, _ := newFakeNode(t, http.StatusTooManyRequests, map[string]string{"Retry-After": "Wed, 21 Oct 2015 07:28:00 GMT"})
	client := newRateLimitedGoEthClient(t, server.URL, tier1(100), "Retry-After", config.RetryAfterHeaderFormatSeconds)

	exceeded := requireExceeded(t, callTwice(t, client))

	assert.Less(t, exceeded.RetryAfter, 5*time.Second)
}

func TestDltIngressRateLimitedGoEthClient_Non429_KeepsBucketAndHttpError(t *testing.T) {
	server, requests := newFakeNode(t, http.StatusInternalServerError, map[string]string{"Retry-After": "30"})
	client := newRateLimitedGoEthClient(t, server.URL, tier1(100), "Retry-After", config.RetryAfterHeaderFormatSeconds)

	err := callTwice(t, client)

	var httpErr rpc.HTTPError
	require.True(t, errors.As(err, &httpErr), "expected rpc.HTTPError, got %T: %v", err, err)
	assert.Equal(t, http.StatusInternalServerError, httpErr.StatusCode)
	assert.False(t, ratelimit.IsExceededError(err))
	assert.Equal(t, int32(2), requests.Load())
}
