//go:build test

package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/ratelimit"
)

func TestDltIngressRateLimitConfig_ToOptionConfigs_RoutesMethodsToTheirBucket(t *testing.T) {
	rateLimit := &RateLimitConfig{Buckets: []RateLimitBucketConfig{
		{Name: "SCARCE", RequestsPerMinute: 1, Methods: []string{"eth_call"}},
		{Name: "REST", RequestsPerMinute: 0, MatchAll: true},
	}}
	limiter := ratelimit.New(rateLimit.ToOptionConfigs()...)
	ctx := context.Background()

	allowed, _ := limiter.Allow(ctx, "eth_call")
	assert.True(t, allowed)
	allowed, retryAfter := limiter.Allow(ctx, "eth_call")
	assert.False(t, allowed)
	assert.Positive(t, retryAfter)

	for range 10 {
		allowed, _ = limiter.Allow(ctx, "eth_getBalance")
		assert.True(t, allowed, "matchAll bucket with 0 rpm is unlimited")
	}
}

func TestDltIngressRateLimitConfig_Validate(t *testing.T) {
	tier := func(name string, methods ...string) RateLimitBucketConfig {
		return RateLimitBucketConfig{Name: name, RequestsPerMinute: 100, Methods: methods}
	}
	matchAll := RateLimitBucketConfig{Name: "DEFAULT", MatchAll: true}

	testCases := []struct {
		name      string
		rateLimit RateLimitConfig
		wantErr   string
	}{
		{
			name:      "tiered scheme",
			rateLimit: RateLimitConfig{Buckets: []RateLimitBucketConfig{tier("T1", "eth_call"), tier("T2", "eth_getBalance")}},
		},
		{
			name:      "flat scheme with header",
			rateLimit: RateLimitConfig{RetryAfterHeader: "Retry-After", RetryAfterHeaderFormat: RetryAfterHeaderFormatUnixTimestamp, Buckets: []RateLimitBucketConfig{matchAll}},
		},
		{
			name:      "tiers with trailing catch-all",
			rateLimit: RateLimitConfig{Buckets: []RateLimitBucketConfig{tier("T1", "eth_call"), matchAll}},
		},
		{
			name:      "format ignored without header",
			rateLimit: RateLimitConfig{RetryAfterHeaderFormat: "bogus", Buckets: []RateLimitBucketConfig{matchAll}},
		},
		{
			name:      "unknown header format",
			rateLimit: RateLimitConfig{RetryAfterHeader: "Retry-After", RetryAfterHeaderFormat: "minutes"},
			wantErr:   "retryAfterHeaderFormat",
		},
		{
			name:      "multiplier stretching the header wait",
			rateLimit: RateLimitConfig{RetryAfterHeader: "Retry-After", RetryAfterHeaderFormat: RetryAfterHeaderFormatSeconds, RetryAfterMultiplier: 1.2},
		},
		{
			name:      "multiplier shrinking the header wait",
			rateLimit: RateLimitConfig{RetryAfterMultiplier: 0.9},
			wantErr:   "retryAfterMultiplier",
		},
		{
			name:      "negative multiplier",
			rateLimit: RateLimitConfig{RetryAfterMultiplier: -1},
			wantErr:   "retryAfterMultiplier",
		},
		{
			name:      "bucket without name",
			rateLimit: RateLimitConfig{Buckets: []RateLimitBucketConfig{tier("", "eth_call")}},
			wantErr:   "has no name",
		},
		{
			name:      "duplicated bucket",
			rateLimit: RateLimitConfig{Buckets: []RateLimitBucketConfig{tier("T1", "eth_call"), tier("T1", "eth_getBalance")}},
			wantErr:   "bucket T1 is duplicated",
		},
		{
			name:      "negative budget",
			rateLimit: RateLimitConfig{Buckets: []RateLimitBucketConfig{{Name: "T1", RequestsPerMinute: -1, Methods: []string{"eth_call"}}}},
			wantErr:   "requestsPerMinute",
		},
		{
			name:      "neither methods nor matchAll",
			rateLimit: RateLimitConfig{Buckets: []RateLimitBucketConfig{tier("T1")}},
			wantErr:   "either methods or matchAll",
		},
		{
			name:      "both methods and matchAll",
			rateLimit: RateLimitConfig{Buckets: []RateLimitBucketConfig{{Name: "T1", Methods: []string{"eth_call"}, MatchAll: true}}},
			wantErr:   "either methods or matchAll",
		},
		{
			name:      "catch-all shadowing later buckets",
			rateLimit: RateLimitConfig{Buckets: []RateLimitBucketConfig{matchAll, tier("T1", "eth_call")}},
			wantErr:   "must be the last one",
		},
		{
			name:      "method in two buckets",
			rateLimit: RateLimitConfig{Buckets: []RateLimitBucketConfig{tier("T1", "eth_call"), tier("T2", "eth_call")}},
			wantErr:   "method eth_call is in buckets T1 and T2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.rateLimit.Validate()
			if tc.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tc.wantErr)
		})
	}
}
