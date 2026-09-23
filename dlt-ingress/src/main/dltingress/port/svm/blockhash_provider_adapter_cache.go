package svm

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/cache"
)

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
