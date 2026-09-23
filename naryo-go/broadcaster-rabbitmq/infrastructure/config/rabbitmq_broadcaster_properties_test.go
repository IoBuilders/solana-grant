//go:build test

package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-rabbitmq/domain/broadcaster"
	coreconfig "gitlab.com/iobuilders/projects/eng/naryo-go/core/infrastructure/config"
)

func validProperties() *RabbitMQBroadcasterProperties {
	return &RabbitMQBroadcasterProperties{
		Host:        "localhost",
		Port:        5672,
		VirtualHost: "/naryo",
		Username:    "guest",
		Password:    "guest",
	}
}

func TestRabbitMQBroadcasterProperties_Map_MapsConnection(t *testing.T) {
	props := validProperties()

	result, err := props.Map()

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "localhost", result.Host)
	assert.Equal(t, 5672, result.Port)
	assert.Equal(t, "/naryo", result.VirtualHost)
	assert.Equal(t, "guest", result.Username)
	assert.Equal(t, "guest", result.Password)
	assert.False(t, result.TLSEnabled)
}

func TestRabbitMQBroadcasterProperties_Map_MapsTLS(t *testing.T) {
	props := validProperties()
	props.TLS = RabbitMQTLSProperties{Enabled: true, Algorithm: "TLSv1.3"}

	result, err := props.Map()

	require.NoError(t, err)
	assert.True(t, result.TLSEnabled)
	assert.Equal(t, broadcaster.TLSAlgorithmTLS13, result.TLSAlgorithm)
}

func TestRabbitMQBroadcasterProperties_Map_TLSEnabledDefaultsAlgorithm(t *testing.T) {
	props := validProperties()
	props.TLS = RabbitMQTLSProperties{Enabled: true}

	result, err := props.Map()

	require.NoError(t, err)
	assert.Equal(t, broadcaster.TLSAlgorithmTLS12, result.TLSAlgorithm)
}

func TestRabbitMQBroadcasterProperties_Map_BlankVirtualHostDefaults(t *testing.T) {
	props := validProperties()
	props.VirtualHost = ""

	result, err := props.Map()

	require.NoError(t, err)
	assert.Equal(t, "/", result.VirtualHost)
}

func TestRabbitMQBroadcasterProperties_Map_InvalidHost(t *testing.T) {
	props := validProperties()
	props.Host = ""

	_, err := props.Map()

	assert.Error(t, err)
}

func TestRabbitMQBroadcasterProperties_Map_InvalidPort(t *testing.T) {
	props := validProperties()
	props.Port = 0

	_, err := props.Map()

	assert.Error(t, err)
}

func TestRabbitMQBroadcasterProperties_Map_NilReceiver(t *testing.T) {
	var props *RabbitMQBroadcasterProperties

	_, err := props.Map()

	assert.EqualError(t, err, "rabbitmq module's configuration not provided")
}

func TestRabbitMQBroadcasterProperties_Map_UnsupportedTLSAlgorithm(t *testing.T) {
	props := validProperties()
	props.TLS = RabbitMQTLSProperties{Enabled: true, Algorithm: "TLSv1.1"}

	_, err := props.Map()

	assert.ErrorContains(t, err, `unsupported TLS algorithm "TLSv1.1"`)
}

// Env-var interpolation is exercised here rather than in the setup package:
// a successful Setup now dials a real broker.
func TestRabbitMQBroadcasterProperties_Map_ResolvesEnvVariables(t *testing.T) {
	t.Setenv("CONFIG_PATH", "")
	t.Setenv("RABBIT_HOST", "broker.internal")
	yml := `
naryo:
  rabbitmq:
    host: ${RABBIT_HOST:localhost}
    port: ${RABBIT_PORT:5672}
    username: ${RABBIT_USER:guest}
    password: ${RABBIT_PASSWORD:guest}
`

	env, err := coreconfig.LoadConfig[RabbitMQEnvironmentProperties](context.Background(), yml, "naryo")
	require.NoError(t, err)

	result, err := env.RabbitMQ.Map()

	require.NoError(t, err)
	assert.Equal(t, "broker.internal", result.Host, "RABBIT_HOST should override the default")
	assert.Equal(t, 5672, result.Port, "unset RABBIT_PORT should fall back to the default")
	assert.Equal(t, "guest", result.Username)
}
