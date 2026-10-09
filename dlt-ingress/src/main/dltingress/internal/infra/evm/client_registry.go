package evm

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/infra/noderatelimit"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/health"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/ratelimit"
)

type ClientRegistry interface {
	GetClientForNetworkId(ctx context.Context, networkId string) (Client, error)
	Shutdown()
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

func (r *ClientRegistryImpl) GetClientForNetworkId(ctx context.Context, networkId string) (Client, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existingClient, found := r.clientsByNetworkId[networkId]; found {
		return existingClient, nil
	}

	return r.registerNewClient(ctx, networkId)
}

func (r *ClientRegistryImpl) Shutdown() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, client := range r.clientsByNetworkId {
		client.Close()
	}
}

func (r *ClientRegistryImpl) registerNewClient(ctx context.Context, networkId string) (Client, error) {
	networkConfig, found := r.networksById[networkId]
	if !found {
		return nil, fmt.Errorf("evm network %s not found in dlt ingress configuration", networkId)
	}

	limiter, found := r.rateLimitersByNetworkId[networkId]
	if !found {
		return nil, fmt.Errorf("no rate limiter configured for network %s", networkId)
	}

	httpClient := &http.Client{Transport: noderatelimit.NewTransport(http.DefaultTransport)}
	goEthClient, err := NewGoEthClient(ctx, networkConfig.Url, httpClient)
	if err != nil {
		return nil, err
	}

	client := NewRateLimitedClient(goEthClient, limiter, networkConfig.RateLimit)
	r.clientsByNetworkId[networkId] = client
	if r.healthRegistry != nil {
		r.healthRegistry.Register(NewNodeChecker(networkId, client))
	}
	return client, nil
}
