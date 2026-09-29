package cache

import (
	"context"
	"fmt"
	"strings"
)

func Cache[T any](ctx context.Context, c Port, fn func() (T, error), key string, keyArgs ...any) (T, error) {
	cacheKey := buildCacheKey(key, keyArgs...)
	// If element exists in the cache, return it
	if value, err := c.Get(ctx, cacheKey); err == nil {
		return value.(T), nil
	}

	// Call function and set value in the cache
	value, err := fn()
	if err != nil {
		return value, err
	}
	if err = c.Set(ctx, cacheKey, value); err != nil {
		var zero T
		return zero, err // Should return nil in this case, because zero value of a pointer is nil
	}
	return value, nil
}

func Evict(ctx context.Context, c Port, fn func() error, key string, keyArgs ...any) error {
	if err := fn(); err != nil {
		return err
	}
	cacheKey := buildCacheKey(key, keyArgs...)
	return c.Delete(ctx, cacheKey)
}

func buildCacheKey(key string, keyArgs ...any) string {
	var sb strings.Builder
	sb.WriteString(key)
	if len(keyArgs) > 0 {
		sb.WriteString(CacheKeySeparator)
		sb.WriteString(fmt.Sprint(keyArgs...))
	}
	return sb.String()
}
