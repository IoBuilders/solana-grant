//go:build test

package custody

import (
	"testing"

	"dlt-ingress/src/main/config"

	"github.com/stretchr/testify/assert"
)

func TestDltIngressNewCustodyProvider_FailsOnInvalidRateLimitConfig(t *testing.T) {
	invalid := &config.RateLimitConfig{Buckets: []config.RateLimitBucketConfig{{Name: "T1"}}}

	_, err := NewCustodyProvider(Config{Provider: string(KMS), Kms: KMSConfig{Region: "us-east-1"}, RateLimit: invalid})

	assert.ErrorContains(t, err, "custody rateLimit")
}
