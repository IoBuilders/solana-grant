//go:build test

package configurationmanager

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/sourceprovider"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/httpclient"
)

// --- stubs ---

type stubHttpClientSourceProvider struct {
	item descriptor.HttpClient
	err  error
}

func (s *stubHttpClientSourceProvider) Load(_ context.Context) (descriptor.HttpClient, error) {
	return s.item, s.err
}

func (s *stubHttpClientSourceProvider) Priority() int { return 1 }

type stubHttpClientDescriptor struct {
	result *httpclient.HttpClient
	err    error
}

func (s *stubHttpClientDescriptor) Map() (*httpclient.HttpClient, error) {
	return s.result, s.err
}

func newHttpClient() *httpclient.HttpClient {
	return httpclient.NewHttpClient(10, time.Second, time.Second, time.Second, time.Second, time.Second, time.Second, false)
}

// --- tests ---

func TestDefaultHttpClientConfigurationManager_Load_Success(t *testing.T) {
	expected := newHttpClient()
	provider := &stubHttpClientSourceProvider{
		item: &stubHttpClientDescriptor{result: expected},
	}
	cm := NewDefaultHttpClientConfigurationManager([]sourceprovider.HttpClientSourceProvider{provider})

	result, err := cm.Load(context.Background())

	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestDefaultHttpClientConfigurationManager_Load_ProviderError(t *testing.T) {
	providerErr := errors.New("source unavailable")
	provider := &stubHttpClientSourceProvider{err: providerErr}
	cm := NewDefaultHttpClientConfigurationManager([]sourceprovider.HttpClientSourceProvider{provider})

	_, err := cm.Load(context.Background())

	assert.ErrorIs(t, err, providerErr)
}

func TestDefaultHttpClientConfigurationManager_Load_DescriptorMapError(t *testing.T) {
	mapErr := errors.New("invalid descriptor")
	provider := &stubHttpClientSourceProvider{
		item: &stubHttpClientDescriptor{err: mapErr},
	}
	cm := NewDefaultHttpClientConfigurationManager([]sourceprovider.HttpClientSourceProvider{provider})

	_, err := cm.Load(context.Background())

	assert.ErrorIs(t, err, mapErr)
}
