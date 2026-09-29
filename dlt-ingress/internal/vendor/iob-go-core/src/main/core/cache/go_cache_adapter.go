package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
)

const DefaultCleanupInterval = 60 * time.Minute
const CacheKeySeparator = "::"

type GoCacheAdapter struct {
	ttls   map[string]time.Duration
	client *cache.Cache
}

func NewGoCache(ttls map[string]time.Duration) *GoCacheAdapter {
	return &GoCacheAdapter{
		ttls:   ttls,
		client: cache.New(cache.NoExpiration, DefaultCleanupInterval),
	}
}

func (c *GoCacheAdapter) Get(ctx context.Context, key string) (any, error) {
	if val, found := c.client.Get(key); found {
		return val, nil
	}
	keyConfigPrefix, _, _ := strings.Cut(key, CacheKeySeparator)
	return nil, fmt.Errorf("key %s not found in the cache", keyConfigPrefix)
}

func (c *GoCacheAdapter) Set(ctx context.Context, key string, value any) error {
	// No expiration if the key is not found in the ttl config
	keyConfigPrefix, _, _ := strings.Cut(key, CacheKeySeparator)
	c.client.Set(key, value, c.ttls[keyConfigPrefix])
	return nil
}

func (c *GoCacheAdapter) Delete(ctx context.Context, key string) error {
	c.client.Delete(key)
	return nil
}
