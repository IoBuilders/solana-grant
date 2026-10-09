//go:build test

package noderatelimit_test

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"testing"
	"time"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/infra/noderatelimit"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/ratelimit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const method = "eth_call"

func newGuard(retryAfterHeaderFormat config.RetryAfterHeaderFormat) *noderatelimit.Guard {
	return newGuardWithMultiplier(retryAfterHeaderFormat, 0)
}

func newGuardWithMultiplier(retryAfterHeaderFormat config.RetryAfterHeaderFormat, retryAfterMultiplier float64) *noderatelimit.Guard {
	limiter := ratelimit.New(ratelimit.WithBucket("TIER_1", 60, ratelimit.MatchAll()))
	return noderatelimit.NewGuard(limiter, &config.RateLimitConfig{
		RetryAfterHeader:       "Retry-After",
		RetryAfterHeaderFormat: retryAfterHeaderFormat,
		RetryAfterMultiplier:   retryAfterMultiplier,
	})
}

// callNode runs one guarded call whose node answers with nodeErr.
func callNode(guard *noderatelimit.Guard, nodeErr error) error {
	_, err := noderatelimit.Call(context.Background(), guard, method, func() (struct{}, error) { return struct{}{}, nodeErr })
	return err
}

func tooManyRequests(retryAfter string) error {
	return &noderatelimit.TooManyRequestsError{Header: http.Header{"Retry-After": []string{retryAfter}}}
}

func assertExceeded(t *testing.T, err error) *ratelimit.ExceededError {
	exceededErr, ok := errors.AsType[*ratelimit.ExceededError](err)
	require.True(t, ok, "expected ratelimit.ExceededError, got %T: %v", err, err)
	return exceededErr
}

func TestDltIngressNodeRateLimitGuard_Call_NonRateLimitErrorKeepsTheBucket(t *testing.T) {
	guard := newGuard(config.RetryAfterHeaderFormatSeconds)
	nodeErr := errors.New("boom")

	assert.Equal(t, nodeErr, callNode(guard, nodeErr))
	assert.NoError(t, callNode(guard, nil))
}

func TestDltIngressNodeRateLimitGuard_Call_429ReturnsExceededError(t *testing.T) {
	guard := newGuard(config.RetryAfterHeaderFormatSeconds)

	err := callNode(guard, tooManyRequests("30"))
	exceededErr := assertExceeded(t, err)

	assert.Equal(t, method, exceededErr.Method)
	assert.Equal(t, 30, exceededErr.RetryAfterSeconds())
}

func TestDltIngressNodeRateLimitGuard_Call_429WithSecondsBlocksForThem(t *testing.T) {
	guard := newGuard(config.RetryAfterHeaderFormatSeconds)

	// First call to notify the exceeded
	errToThrowByNode := tooManyRequests("30")
	_ = callNode(guard, errToThrowByNode)

	// Second call to assert the exceeded status
	err := callNode(guard, nil)
	exceededErr := assertExceeded(t, err)

	assert.Equal(t, 30, exceededErr.RetryAfterSeconds())
}

func TestDltIngressNodeRateLimitGuard_Call_429WithUnixTimestampBlocksUntilIt(t *testing.T) {
	guard := newGuard(config.RetryAfterHeaderFormatUnixTimestamp)

	// First call to notify the exceeded
	errToThrowByNode := tooManyRequests(strconv.FormatInt(time.Now().Add(45*time.Second).Unix(), 10))
	_ = callNode(guard, errToThrowByNode)

	// Second call to assert the exceeded status
	err := callNode(guard, nil)
	exceededErr := assertExceeded(t, err)

	assert.InDelta(t, 45, exceededErr.RetryAfterSeconds(), 1)
}

func TestDltIngressNodeRateLimitGuard_Call_429WithSecondsAndMultiplierBlocksForTheStretchedWait(t *testing.T) {
	guard := newGuardWithMultiplier(config.RetryAfterHeaderFormatSeconds, 1.5)

	// First call to notify the exceeded, already reporting the stretched wait
	err := callNode(guard, tooManyRequests("30"))
	assert.Equal(t, 45, assertExceeded(t, err).RetryAfterSeconds())

	// Second call to assert the bucket stays blocked for the stretched wait too
	err = callNode(guard, nil)
	assert.InDelta(t, 45, assertExceeded(t, err).RetryAfterSeconds(), 1)
}

func TestDltIngressNodeRateLimitGuard_Call_429WithUnixTimestampAndMultiplierBlocksForTheStretchedWait(t *testing.T) {
	guard := newGuardWithMultiplier(config.RetryAfterHeaderFormatUnixTimestamp, 2)

	// First call to notify the exceeded
	errToThrowByNode := tooManyRequests(strconv.FormatInt(time.Now().Add(30*time.Second).Unix(), 10))
	_ = callNode(guard, errToThrowByNode)

	// Second call to assert the exceeded status
	err := callNode(guard, nil)
	exceededErr := assertExceeded(t, err)

	assert.InDelta(t, 60, exceededErr.RetryAfterSeconds(), 2)
}

func TestDltIngressNodeRateLimitGuard_Call_429WithoutConfigStillBlocks(t *testing.T) {
	guard := noderatelimit.NewGuard(ratelimit.New(ratelimit.WithBucket("TIER_1", 60, ratelimit.MatchAll())), nil)

	err := callNode(guard, tooManyRequests("30"))
	assert.Zero(t, assertExceeded(t, err).RetryAfterSeconds())

	err = callNode(guard, nil)
	assertExceeded(t, err)
}

func TestDltIngressNodeRateLimitGuard_Call_429WithUnusableHeaderStillBlocks(t *testing.T) {
	guard := newGuard(config.RetryAfterHeaderFormatSeconds)

	// First call to notify the exceeded, with no reset instant to report
	errToThrowByNode := tooManyRequests("soon")
	err := callNode(guard, errToThrowByNode)
	assert.Zero(t, assertExceeded(t, err).RetryAfterSeconds())

	// Second call to assert the exceeded status
	err = callNode(guard, nil)
	assertExceeded(t, err)
}

func TestDltIngressNodeRateLimitGuard_Call_ExhaustedBucketDoesNotCallTheNode(t *testing.T) {
	guard := newGuard(config.RetryAfterHeaderFormatSeconds)
	_ = callNode(guard, tooManyRequests("30"))

	called := false
	_, err := noderatelimit.Call(context.Background(), guard, method, func() (struct{}, error) {
		called = true
		return struct{}{}, nil
	})

	assertExceeded(t, err)
	assert.False(t, called)
}
