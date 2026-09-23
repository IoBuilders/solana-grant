//go:build test

package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func TestDatabase_NewDatabase(t *testing.T) {
	t.Run("EmptyUrl", func(t *testing.T) {
		_, err := NewDatabase("", 0, 0, 0, 0, 0, 0, 0, 0)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("ZeroOptionalFields_FallBackToDefaults", func(t *testing.T) {
		db, err := NewDatabase("postgres://localhost/db", 0, 0, 0, 0, 0, 0, 0, 0)

		require.NoError(t, err)
		assert.Equal(t, "postgres://localhost/db", db.Url)
		assert.Equal(t, 100, db.MaxOpenConnections)
		assert.Equal(t, 10, db.MaxIdleConnections)
		assert.Equal(t, time.Hour, db.ConnMaxLifetime)
		assert.Equal(t, 30*time.Minute, db.ConnMaxIdleTime)
		assert.Equal(t, 10*time.Second, db.ConnectTimeout)
		assert.Equal(t, 60*time.Second, db.StatementTimeout)
		assert.Equal(t, 60*time.Second, db.IdleInTransactionSessionTimeout)
		assert.Equal(t, 5*time.Second, db.LockTimeout)
	})

	t.Run("NegativeOptionalFields_FallBackToDefaults", func(t *testing.T) {
		db, err := NewDatabase("postgres://localhost/db", -1, -1, -time.Second, -time.Second, -time.Second, -time.Second, -time.Second, -time.Second)

		require.NoError(t, err)
		assert.Equal(t, 100, db.MaxOpenConnections)
		assert.Equal(t, 10, db.MaxIdleConnections)
		assert.Equal(t, time.Hour, db.ConnMaxLifetime)
		assert.Equal(t, 30*time.Minute, db.ConnMaxIdleTime)
		assert.Equal(t, 10*time.Second, db.ConnectTimeout)
		assert.Equal(t, 60*time.Second, db.StatementTimeout)
		assert.Equal(t, 60*time.Second, db.IdleInTransactionSessionTimeout)
		assert.Equal(t, 5*time.Second, db.LockTimeout)
	})

	t.Run("PositiveOptionalFields_OverrideDefaults", func(t *testing.T) {
		db, err := NewDatabase(
			"postgres://localhost/db",
			200,
			20,
			2*time.Hour,
			time.Hour,
			20*time.Second,
			90*time.Second,
			90*time.Second,
			15*time.Second,
		)

		require.NoError(t, err)
		assert.Equal(t, 200, db.MaxOpenConnections)
		assert.Equal(t, 20, db.MaxIdleConnections)
		assert.Equal(t, 2*time.Hour, db.ConnMaxLifetime)
		assert.Equal(t, time.Hour, db.ConnMaxIdleTime)
		assert.Equal(t, 20*time.Second, db.ConnectTimeout)
		assert.Equal(t, 90*time.Second, db.StatementTimeout)
		assert.Equal(t, 90*time.Second, db.IdleInTransactionSessionTimeout)
		assert.Equal(t, 15*time.Second, db.LockTimeout)
	})
}
