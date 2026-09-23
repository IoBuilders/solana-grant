//go:build test

package configurationmanager

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-rabbitmq/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-rabbitmq/app/sourceprovider"
	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-rabbitmq/domain/broadcaster"
)

// --- stubs ---

type stubRabbitMQBroadcasterSourceProvider struct {
	item descriptor.RabbitMQBroadcaster
	err  error
}

func (s *stubRabbitMQBroadcasterSourceProvider) Load(_ context.Context) (descriptor.RabbitMQBroadcaster, error) {
	return s.item, s.err
}

func (s *stubRabbitMQBroadcasterSourceProvider) Priority() int { return 1 }

type stubRabbitMQBroadcasterDescriptor struct {
	result *broadcaster.RabbitMQBroadcaster
	err    error
}

func (s *stubRabbitMQBroadcasterDescriptor) Map() (*broadcaster.RabbitMQBroadcaster, error) {
	return s.result, s.err
}

// --- tests ---

func TestDefaultRabbitMQBroadcasterConfigurationManager_Load_Success(t *testing.T) {
	expected, err := broadcaster.NewRabbitMQBroadcaster("localhost", 5672, "/", "guest", "guest", false, "")
	require.NoError(t, err)
	provider := &stubRabbitMQBroadcasterSourceProvider{
		item: &stubRabbitMQBroadcasterDescriptor{result: expected},
	}
	cm := NewDefaultRabbitMQBroadcasterConfigurationManager([]sourceprovider.RabbitMQBroadcasterSourceProvider{provider})

	result, err := cm.Load(context.Background())

	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestDefaultRabbitMQBroadcasterConfigurationManager_Load_ProviderError(t *testing.T) {
	providerErr := errors.New("source unavailable")
	provider := &stubRabbitMQBroadcasterSourceProvider{err: providerErr}
	cm := NewDefaultRabbitMQBroadcasterConfigurationManager([]sourceprovider.RabbitMQBroadcasterSourceProvider{provider})

	_, err := cm.Load(context.Background())

	assert.ErrorIs(t, err, providerErr)
}

func TestDefaultRabbitMQBroadcasterConfigurationManager_Load_DescriptorMapError(t *testing.T) {
	mapErr := errors.New("invalid descriptor")
	provider := &stubRabbitMQBroadcasterSourceProvider{
		item: &stubRabbitMQBroadcasterDescriptor{err: mapErr},
	}
	cm := NewDefaultRabbitMQBroadcasterConfigurationManager([]sourceprovider.RabbitMQBroadcasterSourceProvider{provider})

	_, err := cm.Load(context.Background())

	assert.ErrorIs(t, err, mapErr)
}
