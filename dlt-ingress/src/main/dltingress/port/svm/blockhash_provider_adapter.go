package svm

import (
	"context"
	"fmt"
)

type BlockhashProviderAdapter struct {
	registry ClientRegistry
}

func NewBlockhashProviderAdapter(registry ClientRegistry) *BlockhashProviderAdapter {
	return &BlockhashProviderAdapter{
		registry: registry,
	}
}

func (p *BlockhashProviderAdapter) GetRecentBlockhash(ctx context.Context, networkId string) (string, error) {
	client, err := p.registry.GetClientForNetworkId(networkId)
	if err != nil {
		return "", fmt.Errorf("error getting solana client for network %s: %w", networkId, err)
	}

	return client.GetRecentBlockhash(ctx)
}
