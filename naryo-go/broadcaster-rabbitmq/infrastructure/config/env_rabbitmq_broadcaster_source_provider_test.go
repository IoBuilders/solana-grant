//go:build test

package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvRabbitMQBroadcasterSourceProvider_Load_ReturnsRabbitMQBroadcasterProperties(t *testing.T) {
	props := validProperties()
	sp := NewEnvRabbitMQBroadcasterSourceProvider(&RabbitMQEnvironmentProperties{RabbitMQ: props})

	result, err := sp.Load(context.Background())

	require.NoError(t, err)
	assert.Equal(t, props, result)
}

func TestEnvRabbitMQBroadcasterSourceProvider_Load_ReturnsNilWhenNotConfigured(t *testing.T) {
	sp := NewEnvRabbitMQBroadcasterSourceProvider(&RabbitMQEnvironmentProperties{RabbitMQ: nil})

	result, err := sp.Load(context.Background())

	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestEnvRabbitMQBroadcasterSourceProvider_Priority(t *testing.T) {
	sp := NewEnvRabbitMQBroadcasterSourceProvider(&RabbitMQEnvironmentProperties{})

	assert.Equal(t, 1, sp.Priority())
}
