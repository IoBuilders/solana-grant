//go:build test

package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvFilterSourceProvider_Load_Empty(t *testing.T) {
	props := &EnvironmentProperties{
		Filters: []*FilterProperties{},
	}
	sp := NewEnvFilterSourceProvider(props)

	result, err := sp.Load(context.Background())

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestEnvFilterSourceProvider_Load_ReturnsAllFilters(t *testing.T) {
	f1 := &FilterProperties{ID: "00000000-0000-0000-0000-000000000001", Name: "filter-one"}
	f2 := &FilterProperties{ID: "00000000-0000-0000-0000-000000000002", Name: "filter-two"}
	props := &EnvironmentProperties{
		Filters: []*FilterProperties{f1, f2},
	}
	sp := NewEnvFilterSourceProvider(props)

	result, err := sp.Load(context.Background())

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, f1, result[0])
	assert.Equal(t, f2, result[1])
}

func TestEnvFilterSourceProvider_Priority(t *testing.T) {
	sp := NewEnvFilterSourceProvider(&EnvironmentProperties{})

	assert.Equal(t, 1, sp.Priority())
}
