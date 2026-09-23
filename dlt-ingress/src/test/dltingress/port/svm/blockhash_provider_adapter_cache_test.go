package mocks

import (
	"context"
	"dlt-ingress/src/main/dltingress/port/svm"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/test/core/cache"
)

const (
	dummyCacheKey = svm.CacheKey + "::" + dummyNetworkId
)

func TestDltIngressBlockhashProviderAdapterCache_GetRecentBlockhash_ReturnsCachedValue(t *testing.T) {
	blockhashProvider := &BlockhashProviderMock{}
	cacheMock := &cache.CachePortMock{}
	ctx := context.Background()

	cacheMock.On("Get", ctx, dummyCacheKey).Return(dummyBlockhash, nil).Once()

	adapter := svm.NewBlockhashProviderAdapterCache(blockhashProvider, cacheMock)

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

	adapter := svm.NewBlockhashProviderAdapterCache(blockhashProvider, cacheMock)

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

	adapter := svm.NewBlockhashProviderAdapterCache(blockhashProvider, cacheMock)

	_, err := adapter.GetRecentBlockhash(ctx, dummyNetworkId)

	assert.Equal(t, errorToThrow, err)
	blockhashProvider.AssertExpectations(t)
	cacheMock.AssertExpectations(t)
}

func TestDltIngressBlockhashProviderAdapterCache_GetRecentBlockhash_UsesSeparateCacheKeyPerNetwork(t *testing.T) {
	secondNetworkId := "solana-devnet"
	secondBlockhash := "ESymwgTNX1j3E4qhKfJAUE41nBWEwXufoYryPbkde5ci"
	secondCacheKey := svm.CacheKey + "::" + secondNetworkId

	blockhashProvider := &BlockhashProviderMock{}
	cacheMock := &cache.CachePortMock{}
	ctx := context.Background()

	cacheMock.On("Get", ctx, dummyCacheKey).Return(dummyBlockhash, nil).Once()
	cacheMock.On("Get", ctx, secondCacheKey).Return(secondBlockhash, nil).Once()

	adapter := svm.NewBlockhashProviderAdapterCache(blockhashProvider, cacheMock)

	result1, err1 := adapter.GetRecentBlockhash(ctx, dummyNetworkId)
	result2, err2 := adapter.GetRecentBlockhash(ctx, secondNetworkId)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, dummyBlockhash, result1)
	assert.Equal(t, secondBlockhash, result2)
	blockhashProvider.AssertNotCalled(t, "GetRecentBlockhash")
	cacheMock.AssertExpectations(t)
}
