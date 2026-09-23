//go:build test

package config

import (
	"context"
	_ "embed"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/base_config.yml
var baseConfigYml string

//go:embed testdata/env_var_config.yml
var envVarConfigYml string

//go:embed testdata/all_properties_config.yml
var allPropertiesConfigYml string

const rootProperty = "naryo"

func TestLoadConfig_BaseYaml_ParsesCorrectly(t *testing.T) {
	t.Setenv("CONFIG_PATH", "")

	cfg, err := LoadConfig[EnvironmentProperties](context.Background(), baseConfigYml, rootProperty)

	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.NotNil(t, cfg.HttpClient)
	assert.Equal(t, 10, cfg.HttpClient.MaxIdleConnections)
	assert.Equal(t, 30*time.Second, cfg.HttpClient.KeepAliveDuration)
	assert.Equal(t, 5*time.Second, cfg.HttpClient.ConnectTimeout)
	assert.Equal(t, 10*time.Second, cfg.HttpClient.ReadTimeout)
	assert.Equal(t, 10*time.Second, cfg.HttpClient.WriteTimeout)
	assert.Equal(t, 15*time.Second, cfg.HttpClient.CallTimeout)
	assert.Equal(t, 20*time.Second, cfg.HttpClient.PingInterval)
	assert.False(t, cfg.HttpClient.RetryOnConnectionFailure)
}

func TestLoadConfig_EmptyYaml_ReturnsEmptyConfig(t *testing.T) {
	t.Setenv("CONFIG_PATH", "")

	cfg, err := LoadConfig[EnvironmentProperties](context.Background(), "", "")

	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Nil(t, cfg.HttpClient)
	assert.Empty(t, cfg.Nodes)
}

func TestLoadConfig_RootPropertyNotFound_ReturnsError(t *testing.T) {
	t.Setenv("CONFIG_PATH", "")

	_, err := LoadConfig[EnvironmentProperties](context.Background(), baseConfigYml, "nonexistent")

	assert.ErrorContains(t, err, "root property")
}

func TestLoadConfig_EnvVar_UsesValueFromEnvironment(t *testing.T) {
	t.Setenv("CONFIG_PATH", "")
	t.Setenv("MAX_IDLE_CONNS", "42")

	cfg, err := LoadConfig[EnvironmentProperties](context.Background(), envVarConfigYml, rootProperty)

	require.NoError(t, err)
	require.NotNil(t, cfg.HttpClient)
	assert.Equal(t, 42, cfg.HttpClient.MaxIdleConnections)
}

func TestLoadConfig_EnvVar_UsesDefaultWhenNotSet(t *testing.T) {
	t.Setenv("CONFIG_PATH", "")
	require.NoError(t, os.Unsetenv("MAX_IDLE_CONNS"))

	cfg, err := LoadConfig[EnvironmentProperties](context.Background(), envVarConfigYml, rootProperty)

	require.NoError(t, err)
	require.NotNil(t, cfg.HttpClient)
	assert.Equal(t, 99, cfg.HttpClient.MaxIdleConnections)
}

func TestLoadConfig_WithOverrideFile_MergesValues(t *testing.T) {
	tmp, err := os.CreateTemp(t.TempDir(), "override_*.yml")
	require.NoError(t, err)
	_, err = tmp.WriteString("naryo:\n  httpClient:\n    maxIdleConnections: 50\n")
	require.NoError(t, err)
	require.NoError(t, tmp.Close())

	t.Setenv("CONFIG_PATH", tmp.Name())

	cfg, err := LoadConfig[EnvironmentProperties](context.Background(), baseConfigYml, rootProperty)

	require.NoError(t, err)
	require.NotNil(t, cfg.HttpClient)
	assert.Equal(t, 50, cfg.HttpClient.MaxIdleConnections)
	assert.Equal(t, 30*time.Second, cfg.HttpClient.KeepAliveDuration)
}

func TestLoadConfig_OverrideFileNotFound_ReturnsError(t *testing.T) {
	t.Setenv("CONFIG_PATH", "/non/existent/path/config.yml")

	_, err := LoadConfig[EnvironmentProperties](context.Background(), baseConfigYml, rootProperty)

	assert.Error(t, err)
}

func TestLoadConfig_OverrideFileUnsupportedExtension_ReturnsError(t *testing.T) {
	tmp, err := os.CreateTemp(t.TempDir(), "override_*.txt")
	require.NoError(t, err)
	require.NoError(t, tmp.Close())

	t.Setenv("CONFIG_PATH", filepath.ToSlash(tmp.Name()))

	_, err = LoadConfig[EnvironmentProperties](context.Background(), baseConfigYml, rootProperty)

	assert.ErrorContains(t, err, "unsupported config file extension")
}

func TestLoadConfig_AllProperties_ParsedCorrectly(t *testing.T) {
	t.Setenv("CONFIG_PATH", "")

	cfg, err := LoadConfig[EnvironmentProperties](context.Background(), allPropertiesConfigYml, rootProperty)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	// httpClient
	require.NotNil(t, cfg.HttpClient)
	assert.Equal(t, 10, cfg.HttpClient.MaxIdleConnections)
	assert.True(t, cfg.HttpClient.RetryOnConnectionFailure)

	// nodes
	require.Len(t, cfg.Nodes, 1)
	assert.Equal(t, "00000000-0000-0000-0000-000000000001", cfg.Nodes[0].ID)
	assert.Equal(t, "solana-node", cfg.Nodes[0].Name)
	assert.Equal(t, "SOLANA", cfg.Nodes[0].Type)
	assert.Equal(t, "WS", cfg.Nodes[0].Connection.Type)
	assert.Equal(t, "ws://localhost:8900", cfg.Nodes[0].Connection.Endpoint.URL)
	assert.Equal(t, "PUBSUB", cfg.Nodes[0].Subscription.Method)
	assert.Equal(t, uint64(100), cfg.Nodes[0].Subscription.InitialSlot)

	// broadcasting.configuration
	require.NotNil(t, cfg.Broadcasting)
	require.Len(t, cfg.Broadcasting.Configuration, 1)
	assert.Equal(t, "00000000-0000-0000-0000-000000000002", cfg.Broadcasting.Configuration[0].ID)
	assert.Equal(t, "HTTP", cfg.Broadcasting.Configuration[0].Type)
	endpoint, ok := cfg.Broadcasting.Configuration[0].AdditionalProperties["endpoint"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "http://localhost:8080", endpoint["url"])

	// broadcasting.broadcasters
	require.Len(t, cfg.Broadcasting.Broadcasters, 1)
	assert.Equal(t, "00000000-0000-0000-0000-000000000003", cfg.Broadcasting.Broadcasters[0].ID)
	assert.Equal(t, "00000000-0000-0000-0000-000000000002", cfg.Broadcasting.Broadcasters[0].ConfigurationID)
	assert.Equal(t, "BLOCK", cfg.Broadcasting.Broadcasters[0].Target.Type)
	assert.Equal(t, []string{"/events"}, cfg.Broadcasting.Broadcasters[0].Target.Destinations)

	// filters
	require.Len(t, cfg.Filters, 1)
	assert.Equal(t, "00000000-0000-0000-0000-000000000004", cfg.Filters[0].ID)
	assert.Equal(t, "tx-filter", cfg.Filters[0].Name)
	assert.Equal(t, "00000000-0000-0000-0000-000000000001", cfg.Filters[0].NodeID)
	assert.Equal(t, "TRANSACTION", cfg.Filters[0].Type)
	assert.Equal(t, "HASH", cfg.Filters[0].IdentifierType)
	assert.Equal(t, []string{"0xabc123"}, cfg.Filters[0].Value)

	// stores
	require.Len(t, cfg.Stores, 1)
	assert.Equal(t, "00000000-0000-0000-0000-000000000001", cfg.Stores[0].NodeID)
	assert.Equal(t, "GORM", cfg.Stores[0].Type)
	require.Len(t, cfg.Stores[0].Features, 2)
	assert.Equal(t, "EVENT", cfg.Stores[0].Features[0].Type)
	assert.Equal(t, "BLOCK_BASED", cfg.Stores[0].Features[0].Strategy)
	require.Len(t, cfg.Stores[0].Features[0].Targets, 1)
	assert.Equal(t, "TRANSACTION", cfg.Stores[0].Features[0].Targets[0].Type)
	assert.Equal(t, "transactions", cfg.Stores[0].Features[0].Targets[0].Destination)
	assert.Equal(t, "FILTER_SYNC", cfg.Stores[0].Features[1].Type)
	assert.Equal(t, "filters", cfg.Stores[0].Features[1].Destination)
}
