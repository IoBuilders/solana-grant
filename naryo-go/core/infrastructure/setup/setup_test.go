//go:build test

package coresetup

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
)

const rootProperty = "naryo"

// allPropertiesYaml exercises all five configuration managers
// SetupConfigManagers wires: httpClient, one node, one broadcaster
// configuration, one broadcaster and one filter.
const allPropertiesYaml = `naryo:
  httpClient:
    maxIdleConnections: 10
    keepAliveDuration: 30s
    connectTimeout: 5s
    readTimeout: 10s
    writeTimeout: 10s
    callTimeout: 15s
    pingInterval: 20s
    retryOnConnectionFailure: true

  nodes:
    - id: "00000000-0000-0000-0000-000000000001"
      name: "solana-node"
      type: "SOLANA"
      connection:
        type: "WS"
        endpoint:
          url: "ws://localhost:8900"
        retry:
          maxRetries: 5
          initialDelay: 1s
          maxDelay: 30s
          multiplier: 2.0
      subscription:
        method: "PUBSUB"
        initialSlot: 100
        interval: 1s

  broadcasting:
    configuration:
      - id: "00000000-0000-0000-0000-000000000002"
        type: "HTTP"
        endpoint:
          url: "http://localhost:8080"
        retry:
          maxRetries: 5
          initialDelay: 1s
          maxDelay: 30s
          multiplier: 2.0
    broadcasters:
      - id: "00000000-0000-0000-0000-000000000003"
        configurationId: "00000000-0000-0000-0000-000000000002"
        target:
          type: "BLOCK"
          destinations:
            - "/events"

  filters:
    - id: "00000000-0000-0000-0000-000000000004"
      name: "tx-filter"
      nodeId: "00000000-0000-0000-0000-000000000001"
      type: "TRANSACTION"
      identifierType: "HASH"
      value:
        - "0xabc123"
      statuses:
        - "FAILED"
`

// resetFlags lets config.LoadConfig re-register its "config" flag on each
// test call without panicking: LoadConfig calls flag.String("config", ...)
// on every invocation, and registering the same flag twice on the
// process-global flag.CommandLine panics.
func resetFlags(t *testing.T) {
	t.Helper()
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
}

// --- SetupConfigManagers ---

func TestSetupConfigManagers_AllProperties_WiresManagersThatLoadCorrectly(t *testing.T) {
	resetFlags(t)
	t.Setenv("CONFIG_PATH", "")

	managers, err := SetupConfigManagers(context.Background(), allPropertiesYaml, rootProperty)

	require.NoError(t, err)
	require.NotNil(t, managers)

	// SetupConfigManagers wires an empty mapper registry; registering the type-specific mapper
	// (here, a passthrough standing in for the one broadcaster-http's own Setup registers) is the
	// responsibility of whichever adapter module owns that broadcaster type — see main.go, which
	// passes managers.BroadcasterConfigConfigManager into broadcasterhttpsetup.Setup.
	err = managers.BroadcasterConfigConfigManager.RegisterMapper(context.Background(), broadcaster.TypeHTTP.String(), func(_ context.Context, source broadcaster.Configuration) (broadcaster.Configuration, error) {
		return source, nil
	})
	require.NoError(t, err)

	httpClient, err := managers.HttpClientConfigManager.Load(context.Background())
	require.NoError(t, err)
	require.NotNil(t, httpClient)
	assert.Equal(t, 10, httpClient.MaxIdleConnections)

	nodes, err := managers.NodeConfigManager.Load(context.Background())
	require.NoError(t, err)
	require.Len(t, nodes, 1)
	assert.Equal(t, "00000000-0000-0000-0000-000000000001", nodes[0].ID.String())
	assert.Equal(t, "solana-node", nodes[0].Name.String())

	broadcasters, err := managers.BroadcasterConfigManager.Load(context.Background())
	require.NoError(t, err)
	require.Len(t, broadcasters, 1)
	assert.Equal(t, "00000000-0000-0000-0000-000000000003", broadcasters[0].ID.String())

	broadcasterConfigs, err := managers.BroadcasterConfigConfigManager.Load(context.Background())
	require.NoError(t, err)
	require.Len(t, broadcasterConfigs, 1)
	assert.Equal(t, "00000000-0000-0000-0000-000000000002", broadcasterConfigs[0].ID().String())

	filters, err := managers.FilterConfigManager.Load(context.Background())
	require.NoError(t, err)
	require.Len(t, filters, 1)
	assert.Equal(t, "00000000-0000-0000-0000-000000000004", filters[0].ID().String())
	assert.Equal(t, "tx-filter", filters[0].Name().String())
}

func TestSetupConfigManagers_LoadConfigError_Propagates(t *testing.T) {
	resetFlags(t)
	t.Setenv("CONFIG_PATH", "")

	managers, err := SetupConfigManagers(context.Background(), allPropertiesYaml, "nonexistent")

	assert.ErrorContains(t, err, "root property")
	assert.Nil(t, managers)
}

// --- Setup ---

func TestSetup_WiresBootstrapperAroundConfigManagers(t *testing.T) {
	configManagers := &ConfigManagers{}

	module, err := Setup(configManagers, nil, nil, nil)

	require.NoError(t, err)
	require.NotNil(t, module)
	assert.Same(t, configManagers, module.ConfigManagers)
	require.NotNil(t, module.Bootstrapper)
	assert.False(t, module.Bootstrapper.IsStarted(context.Background()))
}
