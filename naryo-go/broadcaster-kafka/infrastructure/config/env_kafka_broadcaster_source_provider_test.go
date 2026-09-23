//go:build test

package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvKafkaBroadcasterSourceProvider_Load_ReturnsKafkaBroadcasterProperties(t *testing.T) {
	props := &KafkaBroadcasterProperties{Brokers: []string{"broker1:9092"}}
	sp := NewEnvKafkaBroadcasterSourceProvider(&KafkaEnvironmentProperties{Kafka: props})

	result, err := sp.Load(context.Background())

	require.NoError(t, err)
	assert.Equal(t, props, result)
}

func TestEnvKafkaBroadcasterSourceProvider_Load_ReturnsNilWhenNotConfigured(t *testing.T) {
	sp := NewEnvKafkaBroadcasterSourceProvider(&KafkaEnvironmentProperties{Kafka: nil})

	result, err := sp.Load(context.Background())

	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestEnvKafkaBroadcasterSourceProvider_Priority(t *testing.T) {
	sp := NewEnvKafkaBroadcasterSourceProvider(&KafkaEnvironmentProperties{})

	assert.Equal(t, 1, sp.Priority())
}
