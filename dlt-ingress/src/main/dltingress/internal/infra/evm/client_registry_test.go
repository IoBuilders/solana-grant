//go:build test

package evm_test

import (
	"context"
	"sync"
	"testing"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/infra/evm"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/health"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testNetworkId = "test-network"

func newTestRegistry(t *testing.T) *evm.ClientRegistryImpl {
	return newTestRegistryWithHealth(t, nil)
}

func newTestRegistryWithHealth(t *testing.T, healthRegistry *health.Registry) *evm.ClientRegistryImpl {
	registry, err := evm.NewClientRegistry([]*config.NetworkConfig{
		{Id: testNetworkId, Dlt: "EVM", Url: "http://127.0.0.1:1"},
	}, healthRegistry)
	require.NoError(t, err)
	t.Cleanup(registry.Shutdown)
	return registry
}

func TestDltIngressClientRegistry_GetClientForNetworkId_ConcurrentCallsShareOneClient(t *testing.T) {
	registry := newTestRegistry(t)

	const callers = 32
	clients := make([]evm.Client, callers)
	var wg sync.WaitGroup
	for i := range callers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client, err := registry.GetClientForNetworkId(context.Background(), testNetworkId)
			assert.NoError(t, err)
			clients[i] = client
		}()
	}
	wg.Wait()

	for _, client := range clients {
		require.NotNil(t, client)
		assert.Same(t, clients[0], client)
	}
}

func TestDltIngressClientRegistry_GetClientForNetworkId_NilRateLimitConfig(t *testing.T) {
	registry := newTestRegistry(t)

	client, err := registry.GetClientForNetworkId(context.Background(), testNetworkId)

	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestDltIngressClientRegistry_GetClientForNetworkId_UnknownNetwork(t *testing.T) {
	registry := newTestRegistry(t)

	_, err := registry.GetClientForNetworkId(context.Background(), "unknown")

	assert.Error(t, err)
}

func TestDltIngressNewClientRegistry_FailsOnInvalidRateLimitConfig(t *testing.T) {
	invalid := &config.RateLimitConfig{Buckets: []config.RateLimitBucketConfig{{Name: "T1"}}}

	_, err := evm.NewClientRegistry([]*config.NetworkConfig{{Id: testNetworkId, Dlt: "EVM", RateLimit: invalid}}, nil)

	assert.ErrorContains(t, err, "network test-network rateLimit")
}

func TestDltIngressClientRegistry_GetClientForNetworkId_RegistersNodeHealthChecker(t *testing.T) {
	healthRegistry := health.NewRegistry()
	registry := newTestRegistryWithHealth(t, healthRegistry)
	assert.Empty(t, healthRegistry.RunAllChecks(context.Background()).Checks, "no checker before the client is created")

	_, err := registry.GetClientForNetworkId(context.Background(), testNetworkId)
	require.NoError(t, err)
	_, err = registry.GetClientForNetworkId(context.Background(), testNetworkId)
	require.NoError(t, err)

	checks := healthRegistry.RunAllChecks(context.Background()).Checks
	require.Len(t, checks, 1)
	assert.Contains(t, checks, "evm-node-test-network")
	assert.Equal(t, health.StatusDown, checks["evm-node-test-network"].Status, "the test node is unreachable")
}

func TestDltIngressClientRegistry_GetClientForNetworkId_UnknownNetworkRegistersNoHealthChecker(t *testing.T) {
	healthRegistry := health.NewRegistry()
	registry := newTestRegistryWithHealth(t, healthRegistry)

	_, err := registry.GetClientForNetworkId(context.Background(), "unknown")

	require.Error(t, err)
	assert.Empty(t, healthRegistry.RunAllChecks(context.Background()).Checks)
}
