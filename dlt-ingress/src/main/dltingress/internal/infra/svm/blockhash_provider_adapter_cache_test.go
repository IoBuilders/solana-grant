package svm

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	realcache "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/cache"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/test/core/cache"
)

const (
	dummyCacheKey = CacheKey + "::" + dummyNetworkId
)

func TestDltIngressBlockhashProviderAdapterCache_GetRecentBlockhash_ReturnsCachedValue(t *testing.T) {
	blockhashProvider := &BlockhashProviderMock{}
	cacheMock := &cache.CachePortMock{}
	ctx := context.Background()

	cacheMock.On("Get", ctx, dummyCacheKey).Return(dummyBlockhash, nil).Once()

	adapter := NewBlockhashProviderAdapterCache(blockhashProvider, cacheMock)

	result, err := adapter.GetRecentBlockhash(ctx, dummyNetworkId)

	assert.NoError(t, err)
	assert.Equal(t, dummyBlockhash, result)
	blockhashProvider.AssertNotCalled(t, "GetRecentBlockhash")
	cacheMock.AssertExpectations(t)
}

func TestDltIngressBlockhashProviderAdapterCache_GetRecentBlockhash_FetchesAndSetsOnCacheMiss(t *testing.T) {
	blockhashProvider := &BlockhashProviderMock{}
	cacheMock := &cache.CachePortMock{}
	ctx := context.Background()

	cacheMock.On("Get", ctx, dummyCacheKey).Return(nil, errors.New("cache miss")).Once()
	cacheMock.On("Set", ctx, dummyCacheKey, dummyBlockhash).Return(nil).Once()
	blockhashProvider.On("GetRecentBlockhash", ctx, dummyNetworkId).Return(dummyBlockhash, nil).Once()

	adapter := NewBlockhashProviderAdapterCache(blockhashProvider, cacheMock)

	result, err := adapter.GetRecentBlockhash(ctx, dummyNetworkId)

	assert.NoError(t, err)
	assert.Equal(t, dummyBlockhash, result)
	blockhashProvider.AssertExpectations(t)
	cacheMock.AssertExpectations(t)
}

func TestDltIngressBlockhashProviderAdapterCache_GetRecentBlockhash_ReturnsErrorWhenProviderFails(t *testing.T) {
	blockhashProvider := &BlockhashProviderMock{}
	cacheMock := &cache.CachePortMock{}
	ctx := context.Background()

	errorToThrow := errors.New("rpc error")
	cacheMock.On("Get", ctx, dummyCacheKey).Return(nil, errors.New("cache miss")).Once()
	blockhashProvider.On("GetRecentBlockhash", ctx, dummyNetworkId).Return("", errorToThrow).Once()

	adapter := NewBlockhashProviderAdapterCache(blockhashProvider, cacheMock)

	_, err := adapter.GetRecentBlockhash(ctx, dummyNetworkId)

	assert.Equal(t, errorToThrow, err)
	blockhashProvider.AssertExpectations(t)
	cacheMock.AssertExpectations(t)
}

func TestDltIngressBlockhashProviderAdapterCache_GetRecentBlockhash_UsesSeparateCacheKeyPerNetwork(t *testing.T) {
	secondNetworkId := "solana-devnet"
	secondBlockhash := "ESymwgTNX1j3E4qhKfJAUE41nBWEwXufoYryPbkde5ci"
	secondCacheKey := CacheKey + "::" + secondNetworkId

	blockhashProvider := &BlockhashProviderMock{}
	cacheMock := &cache.CachePortMock{}
	ctx := context.Background()

	cacheMock.On("Get", ctx, dummyCacheKey).Return(dummyBlockhash, nil).Once()
	cacheMock.On("Get", ctx, secondCacheKey).Return(secondBlockhash, nil).Once()

	adapter := NewBlockhashProviderAdapterCache(blockhashProvider, cacheMock)

	result1, err1 := adapter.GetRecentBlockhash(ctx, dummyNetworkId)
	result2, err2 := adapter.GetRecentBlockhash(ctx, secondNetworkId)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, dummyBlockhash, result1)
	assert.Equal(t, secondBlockhash, result2)
	blockhashProvider.AssertNotCalled(t, "GetRecentBlockhash")
	cacheMock.AssertExpectations(t)
}

// This is a regression test for the blockhash cache never expiring when its TTL config is left empty
// (as both the bundled examples and the integration test harness used to do): it exercises the real
// cache.GoCacheAdapter end to end, instead of a cache.Port mock, so it actually proves expiry happens
// rather than just proving this adapter calls Get/Set correctly.
func TestDltIngressBlockhashProviderAdapterCache_GetRecentBlockhash_ExpiresAfterConfiguredTtl(t *testing.T) {
	const ttl = 50 * time.Millisecond
	secondBlockhash := "ESymwgTNX1j3E4qhKfJAUE41nBWEwXufoYryPbkde5ci"

	blockhashProvider := &BlockhashProviderMock{}
	ctx := context.Background()
	goCache := realcache.NewGoCache(map[string]time.Duration{CacheKey: ttl})

	blockhashProvider.On("GetRecentBlockhash", ctx, dummyNetworkId).Return(dummyBlockhash, nil).Once()

	adapter := NewBlockhashProviderAdapterCache(blockhashProvider, goCache)

	result1, err1 := adapter.GetRecentBlockhash(ctx, dummyNetworkId)
	assert.NoError(t, err1)
	assert.Equal(t, dummyBlockhash, result1)

	result2, err2 := adapter.GetRecentBlockhash(ctx, dummyNetworkId)
	assert.NoError(t, err2)
	assert.Equal(t, dummyBlockhash, result2)
	blockhashProvider.AssertExpectations(t)

	time.Sleep(2 * ttl)

	blockhashProvider.On("GetRecentBlockhash", ctx, dummyNetworkId).Return(secondBlockhash, nil).Once()

	result3, err3 := adapter.GetRecentBlockhash(ctx, dummyNetworkId)
	assert.NoError(t, err3)
	assert.Equal(t, secondBlockhash, result3)
	blockhashProvider.AssertExpectations(t)
}

// This proves InvalidateRecentBlockhash forces a refetch on its own, independent of the configured
// TTL: the entry is given an hour to live, nowhere near expiring, yet the explicit invalidation still
// makes the next GetRecentBlockhash call miss and fetch again. This is what callers who've been
// rejected for a stale blockhash (e.g. a BlockhashNotFound simulation failure) rely on to not just
// get the same stale value served back to them.
func TestDltIngressBlockhashProviderAdapterCache_InvalidateRecentBlockhash_ForcesRefetchBeforeTtlExpires(t *testing.T) {
	secondBlockhash := "ESymwgTNX1j3E4qhKfJAUE41nBWEwXufoYryPbkde5ci"

	blockhashProvider := &BlockhashProviderMock{}
	ctx := context.Background()
	goCache := realcache.NewGoCache(map[string]time.Duration{CacheKey: time.Hour})

	blockhashProvider.On("GetRecentBlockhash", ctx, dummyNetworkId).Return(dummyBlockhash, nil).Once()

	adapter := NewBlockhashProviderAdapterCache(blockhashProvider, goCache)

	result1, err1 := adapter.GetRecentBlockhash(ctx, dummyNetworkId)
	assert.NoError(t, err1)
	assert.Equal(t, dummyBlockhash, result1)

	assert.NoError(t, adapter.InvalidateRecentBlockhash(ctx, dummyNetworkId))

	blockhashProvider.On("GetRecentBlockhash", ctx, dummyNetworkId).Return(secondBlockhash, nil).Once()

	result2, err2 := adapter.GetRecentBlockhash(ctx, dummyNetworkId)
	assert.NoError(t, err2)
	assert.Equal(t, secondBlockhash, result2)
	blockhashProvider.AssertExpectations(t)
}
