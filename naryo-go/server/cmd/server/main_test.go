//go:build test

package main

import (
	"context"
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	coreconfig "gitlab.com/iobuilders/projects/eng/naryo-go/core/infrastructure/config"
	"gitlab.com/iobuilders/projects/eng/naryo-go/persistence-gorm/infrastructure/config"
)

func TestDefaultConfigYAML_ParsesCorrectly(t *testing.T) {
	t.Setenv("CONFIG_PATH", "")

	cfg, err := coreconfig.LoadConfig[coreconfig.EnvironmentProperties](context.Background(), defaultConfigYAML, rootProperty)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	require.NotNil(t, cfg.HttpClient)
	assert.Equal(t, 10, cfg.HttpClient.MaxIdleConnections)

	require.Len(t, cfg.Nodes, 1)
	assert.Equal(t, "SOLANA", cfg.Nodes[0].Type)
	assert.Equal(t, "HTTP", cfg.Nodes[0].Connection.Type)
	assert.Equal(t, "POLL", cfg.Nodes[0].Subscription.Method)

	require.NotNil(t, cfg.Broadcasting)
	require.Len(t, cfg.Broadcasting.Configuration, 1)
	assert.Equal(t, "HTTP", cfg.Broadcasting.Configuration[0].Type)
	require.Len(t, cfg.Broadcasting.Broadcasters, 1)
	assert.Equal(t, "BLOCK", cfg.Broadcasting.Broadcasters[0].Target.Type)

	require.Len(t, cfg.Stores, 1)
	assert.Equal(t, "GORM", cfg.Stores[0].Type)
	require.Len(t, cfg.Stores[0].Features, 3)
	assert.Equal(t, "EVENT", cfg.Stores[0].Features[0].Type)
	assert.Equal(t, "FILTER_SYNC", cfg.Stores[0].Features[1].Type)
	assert.Equal(t, "LATEST_BLOCK", cfg.Stores[0].Features[2].Type)
}

func TestDefaultConfigYAML_DatabaseSectionParsesCorrectly(t *testing.T) {
	t.Setenv("CONFIG_PATH", "")

	cfg, err := coreconfig.LoadConfig[config.EnvironmentProperties](context.Background(), defaultConfigYAML, rootProperty)

	require.NoError(t, err)
	require.NotNil(t, cfg.Database)
	assert.Equal(t, "postgres://postgres:postgres@localhost:5432/naryo?sslmode=disable", cfg.Database.Url)
}

func TestDefaultConfigYAML_EnvVarsOverrideDefaults(t *testing.T) {
	t.Setenv("CONFIG_PATH", "")
	t.Setenv("SOLANA_NODE_URL", "http://solana.example.com:8899")
	t.Setenv("BROADCASTER_URL", "http://broadcaster.example.com")
	t.Setenv("DATABASE_URL", "postgres://user:pass@db.example.com:5432/naryo")

	cfg, err := coreconfig.LoadConfig[coreconfig.EnvironmentProperties](context.Background(), defaultConfigYAML, rootProperty)
	require.NoError(t, err)
	require.Len(t, cfg.Nodes, 1)
	assert.Equal(t, "http://solana.example.com:8899", cfg.Nodes[0].Connection.Endpoint.URL)
	require.Len(t, cfg.Broadcasting.Configuration, 1)
	endpoint, ok := cfg.Broadcasting.Configuration[0].AdditionalProperties["endpoint"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "http://broadcaster.example.com", endpoint["url"])

	dbCfg, err := coreconfig.LoadConfig[config.EnvironmentProperties](context.Background(), defaultConfigYAML, rootProperty)
	require.NoError(t, err)
	assert.Equal(t, "postgres://user:pass@db.example.com:5432/naryo", dbCfg.Database.Url)
}

// TestLoadConfig_CalledTwiceInSameProcess_DoesNotPanic guards against a regression: main() calls
// coresetup.SetupConfigManagers and persistencegormsetup.Setup back-to-back in the same process,
// and both go through coreconfig.LoadConfig. LoadConfig itself no longer touches the CLI (only
// CONFIG_PATH, read from the environment, which is stateless), so this should never panic — but
// it's cheap to keep proving it.
func TestLoadConfig_CalledTwiceInSameProcess_DoesNotPanic(t *testing.T) {
	t.Setenv("CONFIG_PATH", "")

	assert.NotPanics(t, func() {
		_, err := coreconfig.LoadConfig[coreconfig.EnvironmentProperties](context.Background(), defaultConfigYAML, rootProperty)
		require.NoError(t, err)

		_, err = coreconfig.LoadConfig[config.EnvironmentProperties](context.Background(), defaultConfigYAML, rootProperty)
		require.NoError(t, err)
	})
}

// resetFlags gives a test a clean flag.CommandLine, so resolveConfigPath can register its
// "-config" flag without colliding with a previous test call.
func resetFlags(t *testing.T) {
	t.Helper()
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
}

func TestResolveConfigPath_FlagGiven_SetsConfigPathEnv(t *testing.T) {
	resetFlags(t)
	t.Setenv("CONFIG_PATH", "")
	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	os.Args = []string{origArgs[0], "-config=/tmp/override.yml"}

	resolveConfigPath()

	assert.Equal(t, "/tmp/override.yml", os.Getenv("CONFIG_PATH"))
}

func TestResolveConfigPath_FlagNotGiven_LeavesConfigPathEnvUntouched(t *testing.T) {
	resetFlags(t)
	t.Setenv("CONFIG_PATH", "/already/set.yml")
	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	os.Args = []string{origArgs[0]}

	resolveConfigPath()

	assert.Equal(t, "/already/set.yml", os.Getenv("CONFIG_PATH"))
}

func TestNewHealthMux_HealthEndpoint_ReturnsOK(t *testing.T) {
	mux := newHealthMux()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
