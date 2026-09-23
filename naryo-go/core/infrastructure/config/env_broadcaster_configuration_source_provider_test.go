//go:build test

package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvBroadcasterConfigurationSourceProvider_Load_Empty(t *testing.T) {
	props := &EnvironmentProperties{
		Broadcasting: &BroadcastingProperties{
			Configuration: []*BroadcasterConfigurationProperties{},
		},
	}
	sp := NewEnvBroadcasterConfigurationSourceProvider(props)

	result, err := sp.Load(context.Background())

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestEnvBroadcasterConfigurationSourceProvider_Load_ReturnsAllConfigurations(t *testing.T) {
	bc1 := &BroadcasterConfigurationProperties{ID: "00000000-0000-0000-0000-000000000001", Type: "HTTP"}
	bc2 := &BroadcasterConfigurationProperties{ID: "00000000-0000-0000-0000-000000000002", Type: "HTTP"}
	props := &EnvironmentProperties{
		Broadcasting: &BroadcastingProperties{
			Configuration: []*BroadcasterConfigurationProperties{bc1, bc2},
		},
	}
	sp := NewEnvBroadcasterConfigurationSourceProvider(props)

	result, err := sp.Load(context.Background())

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, bc1, result[0])
	assert.Equal(t, bc2, result[1])
}

func TestEnvBroadcasterConfigurationSourceProvider_Priority(t *testing.T) {
	sp := NewEnvBroadcasterConfigurationSourceProvider(&EnvironmentProperties{
		Broadcasting: &BroadcastingProperties{},
	})

	assert.Equal(t, 1, sp.Priority())
}
