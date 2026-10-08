package noderatelimit

import (
	"fmt"

	"dlt-ingress/src/main/config"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/ratelimit"
)

func NewRateLimiters(networks []*config.NetworkConfig) (map[string]*ratelimit.RateLimiter, error) {
	rateLimiters := make(map[string]*ratelimit.RateLimiter, len(networks))
	for _, network := range networks {
		if network.RateLimit == nil {
			rateLimiters[network.Id] = ratelimit.New()
			continue
		}
		if err := network.RateLimit.Validate(); err != nil {
			return nil, fmt.Errorf("[dltingress] dltingress network %s rateLimit: %w", network.Id, err)
		}
		rateLimiters[network.Id] = ratelimit.New(network.RateLimit.ToOptionConfigs()...)
	}
	return rateLimiters, nil
}
