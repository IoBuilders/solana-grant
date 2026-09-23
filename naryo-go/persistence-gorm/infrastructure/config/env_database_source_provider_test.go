//go:build test

package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvDatabaseSourceProvider_Load(t *testing.T) {
	t.Run("ReturnsConfiguredDatabaseProperties", func(t *testing.T) {
		properties := &DatabaseProperties{Url: "postgres://localhost/db"}
		sp := NewEnvDatabaseSourceProvider(&EnvironmentProperties{Database: properties})

		db, err := sp.Load(context.Background())

		require.NoError(t, err)
		assert.Same(t, properties, db)
	})

	t.Run("NilDatabase_ReturnsNilWithoutError", func(t *testing.T) {
		sp := NewEnvDatabaseSourceProvider(&EnvironmentProperties{})

		db, err := sp.Load(context.Background())

		require.NoError(t, err)
		assert.Nil(t, db)
	})
}

func TestEnvDatabaseSourceProvider_Priority(t *testing.T) {
	sp := NewEnvDatabaseSourceProvider(&EnvironmentProperties{})

	assert.Equal(t, 1, sp.Priority())
}
