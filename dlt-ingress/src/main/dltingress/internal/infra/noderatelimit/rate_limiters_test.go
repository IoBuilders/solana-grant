//go:build test

package noderatelimit_test

import (
	"context"
	"testing"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/infra/noderatelimit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDltIngressNewRateLimiters_BuildsOneLimiterPerNetwork(t *testing.T) {
	flat := &config.RateLimitConfig{Buckets: []config.RateLimitBucketConfig{{Name: "DEFAULT", RequestsPerMinute: 1, MatchAll: true}}}
	networks := []*config.NetworkConfig{
		{Id: "unlimited"},
		{Id: "evm", Dlt: "EVM", RateLimit: flat},
		{Id: "svm", Dlt: "SVM", RateLimit: flat},
	}

	limiters, err := noderatelimit.NewRateLimiters(networks)

	require.NoError(t, err)
	require.Len(t, limiters, 3)
	ctx := context.Background()
	for range 10 {
		allowed, _ := limiters["unlimited"].Allow(ctx, "getBalance")
		assert.True(t, allowed, "a network without rateLimit is unlimited")
	}
	allowed, _ := limiters["evm"].Allow(ctx, "eth_call")
	assert.True(t, allowed)
	allowed, _ = limiters["svm"].Allow(ctx, "sendTransaction")
	assert.True(t, allowed, "each network has its own budget")
	allowed, _ = limiters["svm"].Allow(ctx, "sendTransaction")
	assert.False(t, allowed)
}

func TestDltIngressNewRateLimiters_FailsOnInvalidConfig(t *testing.T) {
	invalid := &config.RateLimitConfig{Buckets: []config.RateLimitBucketConfig{{Name: "T1"}}}

	_, err := noderatelimit.NewRateLimiters([]*config.NetworkConfig{{Id: "broken", RateLimit: invalid}})

	assert.ErrorContains(t, err, "network broken rateLimit")
}
