//go:build test

package configurationmapperregistry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseConfigurationMapperRegistry_RegisterAndMap(t *testing.T) {
	registry := NewBaseConfigurationMapperRegistry[string]()

	err := registry.Register(context.Background(), "upper", func(_ context.Context, source string) (string, error) {
		return source + "!", nil
	})
	require.NoError(t, err)

	result, err := registry.Map(context.Background(), "upper", "hello")

	require.NoError(t, err)
	assert.Equal(t, "hello!", result)
}

func TestBaseConfigurationMapperRegistry_Map_UnregisteredType(t *testing.T) {
	registry := NewBaseConfigurationMapperRegistry[string]()

	result, err := registry.Map(context.Background(), "missing", "hello")

	assert.Empty(t, result)
	assert.EqualError(t, err, "no mapper registered for configuration type missing")
}

func TestBaseConfigurationMapperRegistry_Map_MultipleTypesMapIndependently(t *testing.T) {
	registry := NewBaseConfigurationMapperRegistry[string]()

	require.NoError(t, registry.Register(context.Background(), "upper", func(_ context.Context, source string) (string, error) {
		return source + "-upper", nil
	}))
	require.NoError(t, registry.Register(context.Background(), "lower", func(_ context.Context, source string) (string, error) {
		return source + "-lower", nil
	}))

	upperResult, err := registry.Map(context.Background(), "upper", "hello")
	require.NoError(t, err)
	assert.Equal(t, "hello-upper", upperResult)

	lowerResult, err := registry.Map(context.Background(), "lower", "hello")
	require.NoError(t, err)
	assert.Equal(t, "hello-lower", lowerResult)
}
