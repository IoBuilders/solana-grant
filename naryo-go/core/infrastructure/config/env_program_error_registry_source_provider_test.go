//go:build test

package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvProgramErrorRegistrySourceProvider_Load_ReturnsProgramErrors(t *testing.T) {
	props := ProgramErrors{
		{ProgramID: "program-1", Codes: []ProgramErrorCodeProperties{{Name: "Unauthorized", Code: 6000}}},
	}
	sp := NewEnvProgramErrorRegistrySourceProvider(&EnvironmentProperties{ProgramErrors: props})

	result, err := sp.Load(context.Background())

	require.NoError(t, err)
	assert.Equal(t, props, result)
}

func TestEnvProgramErrorRegistrySourceProvider_Load_ReturnsNilWhenNotConfigured(t *testing.T) {
	sp := NewEnvProgramErrorRegistrySourceProvider(&EnvironmentProperties{})

	result, err := sp.Load(context.Background())

	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestEnvProgramErrorRegistrySourceProvider_Priority(t *testing.T) {
	sp := NewEnvProgramErrorRegistrySourceProvider(&EnvironmentProperties{})

	assert.Equal(t, 1, sp.Priority())
}
