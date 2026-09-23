//go:build test

package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func TestDatabaseProperties_Map(t *testing.T) {
	t.Run("EmptyUrl_PropagatesValidationError", func(t *testing.T) {
		properties := &DatabaseProperties{}

		db, err := properties.Map()

		assert.ErrorIs(t, err, domainerrors.ErrValidation)
		assert.Nil(t, db)
	})

	t.Run("OnlyUrlSet_RestFallBackToDefaults", func(t *testing.T) {
		properties := &DatabaseProperties{Url: "postgres://localhost/db"}

		db, err := properties.Map()

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

	t.Run("EveryFieldSet_MapsThemAllThrough", func(t *testing.T) {
		properties := &DatabaseProperties{
			Url:                             "postgres://localhost/db",
			MaxOpenConnections:              200,
			MaxIdleConnections:              20,
			ConnMaxLifetime:                 2 * time.Hour,
			ConnMaxIdleTime:                 time.Hour,
			ConnectTimeout:                  20 * time.Second,
			StatementTimeout:                90 * time.Second,
			IdleInTransactionSessionTimeout: 90 * time.Second,
			LockTimeout:                     15 * time.Second,
		}

		db, err := properties.Map()

		require.NoError(t, err)
		assert.Equal(t, "postgres://localhost/db", db.Url)
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
