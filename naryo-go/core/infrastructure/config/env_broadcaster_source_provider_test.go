//go:build test

package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvBroadcasterSourceProvider_Load_Empty(t *testing.T) {
	props := &EnvironmentProperties{
		Broadcasting: &BroadcastingProperties{
			Broadcasters: []*BroadcasterProperties{},
		},
	}
	sp := NewEnvBroadcasterSourceProvider(props)

	result, err := sp.Load(context.Background())

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestEnvBroadcasterSourceProvider_Load_ReturnsAllBroadcasters(t *testing.T) {
	b1 := &BroadcasterProperties{ID: "00000000-0000-0000-0000-000000000001"}
	b2 := &BroadcasterProperties{ID: "00000000-0000-0000-0000-000000000002"}
	props := &EnvironmentProperties{
		Broadcasting: &BroadcastingProperties{
			Broadcasters: []*BroadcasterProperties{b1, b2},
		},
	}
	sp := NewEnvBroadcasterSourceProvider(props)

	result, err := sp.Load(context.Background())

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, b1, result[0])
	assert.Equal(t, b2, result[1])
}

func TestEnvBroadcasterSourceProvider_Priority(t *testing.T) {
	sp := NewEnvBroadcasterSourceProvider(&EnvironmentProperties{
		Broadcasting: &BroadcastingProperties{},
	})

	assert.Equal(t, 1, sp.Priority())
}
