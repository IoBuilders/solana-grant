//go:build test

package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvNodeSourceProvider_Load_Empty(t *testing.T) {
	props := &EnvironmentProperties{
		Nodes: []*NodeProperties{},
	}
	sp := NewEnvNodeSourceProvider(props)

	result, err := sp.Load(context.Background())

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestEnvNodeSourceProvider_Load_ReturnsAllNodes(t *testing.T) {
	n1 := &NodeProperties{ID: "00000000-0000-0000-0000-000000000001", Name: "node-one"}
	n2 := &NodeProperties{ID: "00000000-0000-0000-0000-000000000002", Name: "node-two"}
	props := &EnvironmentProperties{
		Nodes: []*NodeProperties{n1, n2},
	}
	sp := NewEnvNodeSourceProvider(props)

	result, err := sp.Load(context.Background())

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, n1, result[0])
	assert.Equal(t, n2, result[1])
}

func TestEnvNodeSourceProvider_Priority(t *testing.T) {
	sp := NewEnvNodeSourceProvider(&EnvironmentProperties{})

	assert.Equal(t, 1, sp.Priority())
}
