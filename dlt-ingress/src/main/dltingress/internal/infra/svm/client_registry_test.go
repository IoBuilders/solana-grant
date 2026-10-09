//go:build test

package svm_test

import (
	"context"
	"sync"
	"testing"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/infra/svm"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/health"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testNetworkId = "test-network"

func newTestRegistry(t *testing.T) *svm.ClientRegistryImpl {
	return newTestRegistryWithHealth(t, nil)
}

func newTestRegistryWithHealth(t *testing.T, healthRegistry *health.Registry) *svm.ClientRegistryImpl {
	registry, err := svm.NewClientRegistry([]*config.NetworkConfig{
		{Id: testNetworkId, Dlt: "SVM", Url: "http://127.0.0.1:1"},
	}, healthRegistry)
	require.NoError(t, err)
	t.Cleanup(func() { _ = registry.Shutdown() })
	return registry
}

func TestDltIngressSvmClientRegistry_GetClientForNetworkId_ConcurrentCallsShareOneClient(t *testing.T) {
	registry := newTestRegistry(t)

	const callers = 32
	clients := make([]svm.Client, callers)
	var wg sync.WaitGroup
	for i := range callers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client, err := registry.GetClientForNetworkId(testNetworkId)
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

func TestDltIngressSvmClientRegistry_GetClientForNetworkId_NilRateLimitConfig(t *testing.T) {
	registry := newTestRegistry(t)

	client, err := registry.GetClientForNetworkId(testNetworkId)

	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestDltIngressSvmClientRegistry_GetClientForNetworkId_UnknownNetwork(t *testing.T) {
	registry := newTestRegistry(t)

	_, err := registry.GetClientForNetworkId("unknown")

	assert.Error(t, err)
}

func TestDltIngressSvmNewClientRegistry_FailsOnInvalidRateLimitConfig(t *testing.T) {
	invalid := &config.RateLimitConfig{Buckets: []config.RateLimitBucketConfig{{Name: "T1"}}}

	_, err := svm.NewClientRegistry([]*config.NetworkConfig{{Id: testNetworkId, Dlt: "SVM", RateLimit: invalid}}, nil)

	assert.ErrorContains(t, err, "network test-network rateLimit")
}

func TestDltIngressSvmClientRegistry_GetClientForNetworkId_RegistersNodeHealthChecker(t *testing.T) {
	healthRegistry := health.NewRegistry()
	registry := newTestRegistryWithHealth(t, healthRegistry)
	assert.Empty(t, healthRegistry.RunAllChecks(context.Background()).Checks, "no checker before the client is created")

	_, err := registry.GetClientForNetworkId(testNetworkId)
	require.NoError(t, err)
	_, err = registry.GetClientForNetworkId(testNetworkId)
	require.NoError(t, err)

	checks := healthRegistry.RunAllChecks(context.Background()).Checks
	require.Len(t, checks, 1)
	assert.Contains(t, checks, "svm-node-test-network")
	assert.Equal(t, health.StatusDown, checks["svm-node-test-network"].Status, "the test node is unreachable")
}

func TestDltIngressSvmClientRegistry_GetClientForNetworkId_UnknownNetworkRegistersNoHealthChecker(t *testing.T) {
	healthRegistry := health.NewRegistry()
	registry := newTestRegistryWithHealth(t, healthRegistry)

	_, err := registry.GetClientForNetworkId("unknown")

	require.Error(t, err)
	assert.Empty(t, healthRegistry.RunAllChecks(context.Background()).Checks)
}
