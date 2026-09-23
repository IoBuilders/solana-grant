package evm

import (
	"context"
	"dlt-ingress/src/main/config"
)

type ClientRegistry interface {
	GetClientForNetworkId(ctx context.Context, networkId string) (Client, error)
	Shutdown()
}

type ClientRegistryImpl struct {
	clientsByNetworkId map[string]Client
}

func NewClientRegistry() *ClientRegistryImpl {
	return &ClientRegistryImpl{
		clientsByNetworkId: make(map[string]Client),
	}
}

func (r *ClientRegistryImpl) GetClientForNetworkId(ctx context.Context, networkId string) (Client, error) {
	if existingClient, found := r.clientsByNetworkId[networkId]; found {
		return existingClient, nil
	}

	return r.registerNewClient(ctx, networkId)
}

func (r *ClientRegistryImpl) Shutdown() {
	for _, client := range r.clientsByNetworkId {
		client.Close()
	}
}

func (r *ClientRegistryImpl) registerNewClient(ctx context.Context, networkId string) (Client, error) {
	networkConfig, err := config.AppConfig.GetDltIngressNetwork(networkId)
	if err != nil {
		return nil, err
	}

	client, err := NewGoEthClient(ctx, networkConfig.Url)
	if err != nil {
		return nil, err
	}

	r.clientsByNetworkId[networkId] = client
	return client, nil
}
