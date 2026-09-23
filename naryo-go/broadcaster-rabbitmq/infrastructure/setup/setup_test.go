//go:build test

package broadcasterrabbitmqsetup

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-rabbitmq/domain/broadcaster"
	configurationmapperregistry "gitlab.com/iobuilders/projects/eng/naryo-go/core/app/configurationmapper"
	corebroadcaster "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

const rootProperty = "naryo"

// fakeBroadcasterConfigConfigManager is a hand-written test double for
// coreconfigurationmanager.BroadcasterConfigurationConfigurationManager, wrapping a real mapper
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

func TestSetup_RabbitMQNotConfigured_ReturnsError(t *testing.T) {
	t.Setenv("CONFIG_PATH", "")
	yml := `
naryo:
  dummyKey: dummyValue
`

	module, err := Setup(context.Background(), yml, rootProperty, newFakeBroadcasterConfigConfigManager())

	assert.ErrorContains(t, err, "rabbitmq module's configuration not provided")
	assert.Nil(t, module)
}

func TestSetup_MissingHost_ReturnsError(t *testing.T) {
	t.Setenv("CONFIG_PATH", "")
	yml := `
naryo:
  rabbitmq:
    port: 5672
    username: guest
    password: guest
`

	module, err := Setup(context.Background(), yml, rootProperty, newFakeBroadcasterConfigConfigManager())

	assert.ErrorContains(t, err, "Host")
	assert.Nil(t, module)
}

func TestSetup_InvalidPort_ReturnsError(t *testing.T) {
	t.Setenv("CONFIG_PATH", "")
	yml := `
naryo:
  rabbitmq:
    host: localhost
    port: 70000
    username: guest
    password: guest
`

	module, err := Setup(context.Background(), yml, rootProperty, newFakeBroadcasterConfigConfigManager())

	assert.ErrorContains(t, err, "port must be between 1 and 65535")
	assert.Nil(t, module)
}

func TestSetup_RootPropertyNotFound_ReturnsError(t *testing.T) {
	t.Setenv("CONFIG_PATH", "")
	yml := `
naryo:
  rabbitmq:
    host: localhost
    port: 5672
    username: guest
    password: guest
`

	module, err := Setup(context.Background(), yml, "nonexistent", newFakeBroadcasterConfigConfigManager())

	assert.ErrorContains(t, err, "root property")
	assert.Nil(t, module)
}

func TestSetup_TLSAlgorithmWithoutEnabled_ReturnsError(t *testing.T) {
	t.Setenv("CONFIG_PATH", "")
	yml := `
naryo:
  rabbitmq:
    host: localhost
    port: 5672
    username: guest
    password: guest
    tls:
      algorithm: TLSv1.3
`

	module, err := Setup(context.Background(), yml, rootProperty, newFakeBroadcasterConfigConfigManager())

	assert.ErrorContains(t, err, "TLS algorithm must not be set while TLS is disabled")
	assert.Nil(t, module)
}

// TestSetup_RegistersRabbitMQConfigurationMapper asserts Setup wires a RABBITMQ mapper into the
// shared registry before it dials the broker, so registration succeeds independently of broker
// connectivity (port 1 refuses connections).
func TestSetup_RegistersRabbitMQConfigurationMapper(t *testing.T) {
	t.Setenv("CONFIG_PATH", "")
	yml := `
naryo:
  rabbitmq:
    host: 127.0.0.1
    port: 1
    username: guest
    password: guest
`
	manager := newFakeBroadcasterConfigConfigManager()

	_, err := Setup(context.Background(), yml, rootProperty, manager)
	require.ErrorContains(t, err, "connecting to rabbitmq")

	t.Run("MapsDestinationExchange", func(t *testing.T) {
		source, err := corebroadcaster.NewGenericConfiguration(uuid.New(), broadcaster.TypeRabbitMQ, map[string]interface{}{
			"destination": map[string]interface{}{"exchange": "naryo-events"},
		})
		require.NoError(t, err)

		mapped, err := manager.registry.Map(context.Background(), broadcaster.TypeRabbitMQ.String(), source)

		require.NoError(t, err)
		rabbitMQConfiguration, ok := mapped.(*broadcaster.RabbitMQConfiguration)
		require.True(t, ok)
		assert.Equal(t, source.ID(), rabbitMQConfiguration.ID())
		assert.Equal(t, broadcaster.Exchange("naryo-events"), rabbitMQConfiguration.Exchange)
	})

	t.Run("MissingExchange", func(t *testing.T) {
		source, err := corebroadcaster.NewGenericConfiguration(uuid.New(), broadcaster.TypeRabbitMQ, nil)
		require.NoError(t, err)

		mapped, err := manager.registry.Map(context.Background(), broadcaster.TypeRabbitMQ.String(), source)

		assert.ErrorIs(t, err, domainerrors.ErrValidation)
		assert.Nil(t, mapped)
	})
}
