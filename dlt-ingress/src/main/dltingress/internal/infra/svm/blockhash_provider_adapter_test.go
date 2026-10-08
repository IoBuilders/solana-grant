package svm

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	dummyNetworkId = "solana-mainnet"
	dummyBlockhash = "CSymwgTNX1j3E4qhKfJAUE41nBWEwXufoYryPbkde5ci"
)

func TestDltIngressBlockhashProviderAdapter_GetRecentBlockhash_Success(t *testing.T) {
	registry := &SvmClientRegistryMock{}
	client := &SvmClientMock{}
	ctx := context.Background()

	client.On("GetRecentBlockhash", ctx).Return(dummyBlockhash, nil).Once()
	registry.On("GetClientForNetworkId", dummyNetworkId).Return(client, nil).Once()

	adapter := NewBlockhashProviderAdapter(registry)

	result, err := adapter.GetRecentBlockhash(ctx, dummyNetworkId)

	assert.NoError(t, err)
	assert.Equal(t, dummyBlockhash, result)
	registry.AssertExpectations(t)
	client.AssertExpectations(t)
}

func TestDltIngressBlockhashProviderAdapter_GetRecentBlockhash_ReturnsErrorWhenRegistryFails(t *testing.T) {
	registry := &SvmClientRegistryMock{}
	ctx := context.Background()

	errorToThrow := errors.New("network not found")
	registry.On("GetClientForNetworkId", dummyNetworkId).Return(nil, errorToThrow).Once()

	adapter := NewBlockhashProviderAdapter(registry)

	_, err := adapter.GetRecentBlockhash(ctx, dummyNetworkId)

	expectedError := fmt.Errorf("error getting solana client for network %s: %w", dummyNetworkId, errorToThrow)
	assert.Equal(t, expectedError, err)
	registry.AssertExpectations(t)
}

func TestDltIngressBlockhashProviderAdapter_GetRecentBlockhash_ReturnsErrorWhenClientFails(t *testing.T) {
	registry := &SvmClientRegistryMock{}
	client := &SvmClientMock{}
	ctx := context.Background()

	errorToThrow := errors.New("rpc error")
	client.On("GetRecentBlockhash", ctx).Return("", errorToThrow).Once()
	registry.On("GetClientForNetworkId", dummyNetworkId).Return(client, nil).Once()

	adapter := NewBlockhashProviderAdapter(registry)

	_, err := adapter.GetRecentBlockhash(ctx, dummyNetworkId)

	assert.Equal(t, errorToThrow, err)
	registry.AssertExpectations(t)
	client.AssertExpectations(t)
}
