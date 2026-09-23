//go:build test

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKafkaBroadcasterProperties_Map_MapsBrokers(t *testing.T) {
	props := &KafkaBroadcasterProperties{Brokers: []string{"broker1:9092", "broker2:9092"}}

	result, err := props.Map()

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, props.Brokers, result.Brokers)
}

func TestKafkaBroadcasterProperties_Map_InvalidBrokers(t *testing.T) {
	props := &KafkaBroadcasterProperties{Brokers: nil}

	_, err := props.Map()

	assert.Error(t, err)
}

func TestKafkaBroadcasterProperties_Map_NilReceiver(t *testing.T) {
	var props *KafkaBroadcasterProperties

	_, err := props.Map()

	assert.EqualError(t, err, "kafka module's configuration not provided")
}
