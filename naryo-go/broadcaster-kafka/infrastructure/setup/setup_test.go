//go:build test

package broadcasterkafkasetup

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-kafka/domain/broadcaster"
	configurationmapperregistry "gitlab.com/iobuilders/projects/eng/naryo-go/core/app/configurationmapper"
	corebroadcaster "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
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

func TestSetup_KafkaNotConfigured_ReturnsError(t *testing.T) {
	t.Setenv("CONFIG_PATH", "")
	yml := `
naryo:
  dummyKey: dummyValue
`
	manager := newFakeBroadcasterConfigConfigManager()

	module, err := Setup(context.Background(), yml, rootProperty, manager)

	assert.ErrorContains(t, err, "kafka module's configuration not provided")
	assert.Nil(t, module)
}

func TestSetup_InvalidBrokers_ReturnsError(t *testing.T) {
	t.Setenv("CONFIG_PATH", "")
	yml := `
naryo:
  kafka:
    brokers: []
`
	manager := newFakeBroadcasterConfigConfigManager()

	module, err := Setup(context.Background(), yml, rootProperty, manager)

	assert.Error(t, err)
	assert.Nil(t, module)
}

func TestSetup_RootPropertyNotFound_ReturnsError(t *testing.T) {
	t.Setenv("CONFIG_PATH", "")
	yml := `
naryo:
  kafka:
    brokers:
      - broker1:9092
`
	manager := newFakeBroadcasterConfigConfigManager()

	module, err := Setup(context.Background(), yml, "nonexistent", manager)

	assert.ErrorContains(t, err, "root property")
	assert.Nil(t, module)
}

// TestSetup_RegistersKafkaConfigurationMapper asserts Setup wires a passthrough
// KAFKA mapper into the shared registry (broadcaster-kafka's own domain type carries
// no adapter-specific shape beyond the generic Configuration, so the mapper is identity)
// before it attempts to dial the actual Kafka brokers, so registration succeeds
// independently of broker connectivity.
func TestSetup_RegistersKafkaConfigurationMapper(t *testing.T) {
	t.Setenv("CONFIG_PATH", "")
	yml := `
naryo:
  kafka:
    brokers:
      - broker1:9092
`
	manager := newFakeBroadcasterConfigConfigManager()

	_, _ = Setup(context.Background(), yml, rootProperty, manager)

	source, err := corebroadcaster.NewGenericConfiguration(uuid.New(), broadcaster.TypeKafka, nil)
	require.NoError(t, err)

	mapped, err := manager.registry.Map(context.Background(), broadcaster.TypeKafka.String(), source)

	require.NoError(t, err)
	assert.Same(t, source, mapped)
}
