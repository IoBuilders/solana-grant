package svm

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/cache"
)

// CacheKey is the name this module registers with the injected cache.Port. Whoever builds that
// cache.Port (CoreDependencies.Cache) must configure a TTL for this key well under Solana's blockhash
// validity window (~60-90s, i.e. ~150 blocks) — without one, go-cache's zero-value default is
// NoExpiration, and a long-running instance ends up signing and sending transactions against a
// blockhash the cluster has long since forgotten.
const CacheKey = "blockhash"

type BlockhashProviderAdapterCache struct {
	blockhashProvider BlockhashProvider
	cache             cache.Port
}

func NewBlockhashProviderAdapterCache(blockhashProvider BlockhashProvider, cache cache.Port) *BlockhashProviderAdapterCache {
	return &BlockhashProviderAdapterCache{
		blockhashProvider: blockhashProvider,
		cache:             cache,
	}
}

func (p *BlockhashProviderAdapterCache) GetRecentBlockhash(ctx context.Context, networkId string) (string, error) {
	return cache.Cache[string](ctx, p.cache, func() (string, error) {
		return p.blockhashProvider.GetRecentBlockhash(ctx, networkId)
	}, CacheKey, networkId)
}

func (p *BlockhashProviderAdapterCache) InvalidateRecentBlockhash(ctx context.Context, networkId string) error {
	return cache.Evict(ctx, p.cache, func() error { return nil }, CacheKey, networkId)
}
