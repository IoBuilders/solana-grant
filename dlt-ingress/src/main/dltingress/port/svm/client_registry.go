package svm

import (
	"dlt-ingress/src/main/config"
)

type ClientRegistry interface {
	GetClientForNetworkId(networkId string) (Client, error)
	Shutdown() error
}

type ClientRegistryImpl struct {
	clientsByNetworkId map[string]Client
}

func NewClientRegistry() *ClientRegistryImpl {
	return &ClientRegistryImpl{
		clientsByNetworkId: make(map[string]Client),
	}
}

func (r *ClientRegistryImpl) GetClientForNetworkId(networkId string) (Client, error) {
	if existingClient, found := r.clientsByNetworkId[networkId]; found {
		return existingClient, nil
	}

	return r.registerNewClient(networkId)
}

func (r *ClientRegistryImpl) Shutdown() error {
	for _, client := range r.clientsByNetworkId {
		if err := client.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (r *ClientRegistryImpl) registerNewClient(networkId string) (Client, error) {
	networkConfig, err := config.AppConfig.GetDltIngressNetwork(networkId)
	if err != nil {
		return nil, err
	}

	client := NewSolanaGoClient(networkConfig.Url)
	r.clientsByNetworkId[networkId] = client
	return client, nil
}
