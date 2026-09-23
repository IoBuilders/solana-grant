//go:build test

package broadcasterhttpsetup

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-http/domain/broadcaster"
	configurationmapperregistry "gitlab.com/iobuilders/projects/eng/naryo-go/core/app/configurationmapper"
	corebroadcaster "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/httpclient"
)

// --- stubs ---

// fakeHttpClientConfigManager is a hand-written test double for
// configurationmanager.HttpClientConfigurationManager.
type fakeHttpClientConfigManager struct {
	result *httpclient.HttpClient
	err    error
}

func (m *fakeHttpClientConfigManager) Load(context.Context) (*httpclient.HttpClient, error) {
	return m.result, m.err
}

// fakeBroadcasterConfigConfigManager is a hand-written test double for
// configurationmanager.BroadcasterConfigurationConfigurationManager, wrapping a real mapper
// registry so tests can assert on what Setup registers via RegisterMapper.
type fakeBroadcasterConfigConfigManager struct {
	registry configurationmapperregistry.ConfigurationMapperRegistry[corebroadcaster.Configuration]
}

func newFakeBroadcasterConfigConfigManager() *fakeBroadcasterConfigConfigManager {
	return &fakeBroadcasterConfigConfigManager{
		registry: configurationmapperregistry.NewBaseConfigurationMapperRegistry[corebroadcaster.Configuration](),
	}
}

func (m *fakeBroadcasterConfigConfigManager) Load(context.Context) ([]corebroadcaster.Configuration, error) {
	return nil, nil
}

func (m *fakeBroadcasterConfigConfigManager) RegisterMapper(ctx context.Context, configurationType string, mapper func(ctx context.Context, source corebroadcaster.Configuration) (corebroadcaster.Configuration, error)) error {
	return m.registry.Register(ctx, configurationType, mapper)
}

// --- tests ---

func TestSetup_Success_WrapsLoadedHttpClient(t *testing.T) {
	httpClient := &httpclient.HttpClient{
		MaxIdleConnections: 1,
		KeepAliveDuration:  time.Second,
		ConnectTimeout:     time.Second,
		ReadTimeout:        time.Second,
	}
	configManager := &fakeHttpClientConfigManager{result: httpClient}
	manager := newFakeBroadcasterConfigConfigManager()

	module, err := Setup(context.Background(), configManager, manager)

	require.NoError(t, err)
	require.NotNil(t, module)
	assert.NotNil(t, module.HttpBroadcasterProducer)
}

func TestSetup_HttpClientLoadError_Propagates(t *testing.T) {
	loadErr := errors.New("boom")
	configManager := &fakeHttpClientConfigManager{err: loadErr}
	manager := newFakeBroadcasterConfigConfigManager()

	module, err := Setup(context.Background(), configManager, manager)

	assert.ErrorIs(t, err, loadErr)
	assert.Nil(t, module)
}

// TestSetup_RegistersHTTPConfigurationMapper asserts Setup wires a working HTTP
// mapper into the shared registry, so callers loading a raw broadcaster.Configuration
// of type HTTP get back a usable *broadcaster.HTTPConfiguration.
func TestSetup_RegistersHTTPConfigurationMapper(t *testing.T) {
	httpClient := &httpclient.HttpClient{
		MaxIdleConnections: 1,
		KeepAliveDuration:  time.Second,
		ConnectTimeout:     time.Second,
		ReadTimeout:        time.Second,
	}
	configManager := &fakeHttpClientConfigManager{result: httpClient}
	manager := newFakeBroadcasterConfigConfigManager()

	_, err := Setup(context.Background(), configManager, manager)
	require.NoError(t, err)

	source, err := corebroadcaster.NewGenericConfiguration(
		uuid.New(),
		broadcaster.TypeHTTP,
		map[string]interface{}{
			"endpoint": map[string]interface{}{"url": "http://localhost:8080"},
			"retry": map[string]interface{}{
				"maxRetries":   0,
				"initialDelay": "1ms",
				"maxDelay":     "1ms",
				"multiplier":   1.0,
			},
		},
	)
	require.NoError(t, err)

	mapped, err := manager.registry.Map(context.Background(), broadcaster.TypeHTTP.String(), source)

	require.NoError(t, err)
	httpConfig, ok := mapped.(*broadcaster.HTTPConfiguration)
	require.True(t, ok)
	assert.Equal(t, "http://localhost:8080", httpConfig.Connection.Endpoint.URL())
}

func TestSetup_RegistersHTTPConfigurationMapper_UnmappableSource_Errors(t *testing.T) {
	httpClient := &httpclient.HttpClient{
		MaxIdleConnections: 1,
		KeepAliveDuration:  time.Second,
		ConnectTimeout:     time.Second,
		ReadTimeout:        time.Second,
	}
	configManager := &fakeHttpClientConfigManager{result: httpClient}
	manager := newFakeBroadcasterConfigConfigManager()

	_, err := Setup(context.Background(), configManager, manager)
	require.NoError(t, err)

	source, err := corebroadcaster.NewGenericConfiguration(uuid.New(), broadcaster.TypeHTTP, nil)
	require.NoError(t, err)

	_, err = manager.registry.Map(context.Background(), broadcaster.TypeHTTP.String(), source)

	assert.Error(t, err)
}
