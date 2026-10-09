package svm

import (
	"fmt"
	"sync"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/infra/noderatelimit"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/health"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/ratelimit"
)

type ClientRegistry interface {
	GetClientForNetworkId(networkId string) (Client, error)
	Shutdown() error
}

type ClientRegistryImpl struct {
	mu                      sync.Mutex
	networksById            map[string]*config.NetworkConfig
	clientsByNetworkId      map[string]Client
	rateLimitersByNetworkId map[string]*ratelimit.RateLimiter
	healthRegistry          *health.Registry
}

func NewClientRegistry(networks []*config.NetworkConfig, healthRegistry *health.Registry) (*ClientRegistryImpl, error) {
	rateLimitersByNetworkId, err := noderatelimit.NewRateLimiters(networks)
	if err != nil {
		return nil, err
	}

	networksById := make(map[string]*config.NetworkConfig, len(networks))
	for _, network := range networks {
		networksById[network.Id] = network
	}
	return &ClientRegistryImpl{
		networksById:            networksById,
		clientsByNetworkId:      make(map[string]Client),
		rateLimitersByNetworkId: rateLimitersByNetworkId,
		healthRegistry:          healthRegistry,
	}, nil
}

func (r *ClientRegistryImpl) GetClientForNetworkId(networkId string) (Client, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existingClient, found := r.clientsByNetworkId[networkId]; found {
		return existingClient, nil
	}

	return r.registerNewClient(networkId)
}

func (r *ClientRegistryImpl) Shutdown() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, client := range r.clientsByNetworkId {
		if err := client.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (r *ClientRegistryImpl) registerNewClient(networkId string) (Client, error) {
	networkConfig, found := r.networksById[networkId]
	if !found {
		return nil, fmt.Errorf("svm network %s not found in dlt ingress configuration", networkId)
	}

	limiter, found := r.rateLimitersByNetworkId[networkId]
	if !found {
		return nil, fmt.Errorf("no rate limiter configured for network %s", networkId)
	}

	client := NewRateLimitedClient(NewSolanaGoClient(networkConfig.Url), limiter, networkConfig.RateLimit)
	r.clientsByNetworkId[networkId] = client
	if r.healthRegistry != nil {
		r.healthRegistry.Register(NewNodeChecker(networkId, client))
	}
	return client, nil
}
