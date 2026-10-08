package custody

import (
	"fmt"
	"strings"
	"time"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/infra/noderatelimit"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/ratelimit"
)

type ProviderName string

const (
	Dfns ProviderName = "dfns"
	KMS  ProviderName = "kms"
)

// DefaultRequestTimeout bounds each custody provider call when no request timeout is configured.
const DefaultRequestTimeout = 30 * time.Second

type Config struct {
	Provider  string
	Dfns      DfnsConfig
	Kms       KMSConfig
	RateLimit *config.RateLimitConfig
	// RequestTimeout bounds each call to the provider, whichever it is. Zero means DefaultRequestTimeout.
	RequestTimeout time.Duration
}

func NewCustodyProvider(cfg Config) (Port, error) {
	guard, err := newRateLimitGuard(cfg.RateLimit)
	if err != nil {
		return nil, err
	}

	requestTimeout := cfg.RequestTimeout
	if requestTimeout <= 0 {
		requestTimeout = DefaultRequestTimeout
	}

	switch ProviderName(strings.ToLower(cfg.Provider)) {
	case Dfns:
		return newDfnsAdapter(cfg.Dfns, guard, requestTimeout)
	case KMS:
		return newKMSAdapter(cfg.Kms, guard, requestTimeout)
	default:
		return nil, fmt.Errorf("unsupported custody provider: %s", cfg.Provider)
	}
}

func newRateLimitGuard(rateLimitConfig *config.RateLimitConfig) (*noderatelimit.Guard, error) {
	if rateLimitConfig == nil {
		return noderatelimit.NewGuard(ratelimit.New(), nil), nil
	}
	if err := rateLimitConfig.Validate(); err != nil {
		return nil, fmt.Errorf("[dltingress] dltingress custody rateLimit: %w", err)
	}
	limiter := ratelimit.New(rateLimitConfig.ToOptionConfigs()...)
	return noderatelimit.NewGuard(limiter, rateLimitConfig), nil
}

func newDfnsAdapter(cfg DfnsConfig, guard *noderatelimit.Guard, requestTimeout time.Duration) (Port, error) {
	client, err := cfg.NewDfnsClient(requestTimeout)
	if err != nil {
		return nil, err
	}
	return NewDfnsProvider(client, guard), nil
}
