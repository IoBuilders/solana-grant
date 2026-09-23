//go:build test

package eventstore

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newMockedLatestBlockStore(t *testing.T) (*SolanaLatestBlockStore, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	require.NoError(t, err)

	mock.MatchExpectationsInOrder(false)

	return NewSolanaLatestBlockStore(gormDB), mock
}

// insertOnConflictNodeID matches the upsert statement Save issues: a plain
// insert whose ON CONFLICT target is node_id, so repeated saves for the same
// node update the existing snapshot row instead of piling up new ones.
const insertOnConflictNodeID = `INSERT INTO "solana_latest_blocks".*ON CONFLICT \("node_id"\) DO UPDATE`

func TestSolanaLatestBlockStore_Save(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		s, mock := newMockedLatestBlockStore(t)
		mock.ExpectBegin()
		// A single upserted row: the latest block is a one-row-per-node
		// snapshot, it has no associations to cascade into.
		mock.ExpectExec(insertOnConflictNodeID).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := s.Save(context.Background(), newSolanaBlockEvent(t))

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("SuccessWithoutBlockTime", func(t *testing.T) {
		s, mock := newMockedLatestBlockStore(t)
		mock.ExpectBegin()
		mock.ExpectExec(insertOnConflictNodeID).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		blockEvent := newSolanaBlockEvent(t, func(f *solanaBlockEventFixture) {
			f.blockTime = nil
		})

		err := s.Save(context.Background(), blockEvent)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("IgnoresBlockTransactions", func(t *testing.T) {
		s, mock := newMockedLatestBlockStore(t)
		nodeID := uuid.New()
		mock.ExpectBegin()
		// Only the snapshot row: the block's transactions are persisted by
		// SolanaBlockEventStore, never by this one.
		mock.ExpectExec(insertOnConflictNodeID).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		blockEvent := newSolanaBlockEvent(t, func(f *solanaBlockEventFixture) {
			f.nodeID = nodeID
		}, withTransaction(nodeID, 42))

		err := s.Save(context.Background(), blockEvent)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("WrapsPersistenceFailure", func(t *testing.T) {
		s, mock := newMockedLatestBlockStore(t)
		dbErr := errors.New("connection reset")
		mock.ExpectBegin()
		mock.ExpectExec(insertOnConflictNodeID).WillReturnError(dbErr)
		mock.ExpectRollback()

		err := s.Save(context.Background(), newSolanaBlockEvent(t))

		require.Error(t, err)
		assert.ErrorIs(t, err, dbErr)
		assert.ErrorContains(t, err, "gorm solana latest block store:")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSolanaLatestBlockStore_Get(t *testing.T) {
	columns := []string{"id", "node_id", "slot", "blockhash", "block_time", "created_at", "updated_at"}

	t.Run("Success", func(t *testing.T) {
		s, mock := newMockedLatestBlockStore(t)
		nodeID := uuid.New()
		now := time.Now()
		mock.ExpectQuery(`SELECT \* FROM "solana_latest_blocks" WHERE node_id.*LIMIT`).WithArgs(nodeID, sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows(columns).
				AddRow(uuid.New(), nodeID, uint64(99), "5VERv8NMvzbJMEkV8xnrLkEaWRtSz9CosKDYjCJjBRnbJ", nil, now, now))

		slot, err := s.Get(context.Background(), nodeID)

		require.NoError(t, err)
		assert.Equal(t, uint64(99), slot)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("FiltersByNode", func(t *testing.T) {
		// Two nodes each have their own snapshot row; Get must only ever
		// query for the requested node, not just the first row it finds.
		s, mock := newMockedLatestBlockStore(t)
		nodeID := uuid.New()
		otherNodeID := uuid.New()
		now := time.Now()
		mock.ExpectQuery(`SELECT \* FROM "solana_latest_blocks" WHERE node_id.*LIMIT`).WithArgs(nodeID, sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows(columns).
				AddRow(uuid.New(), nodeID, uint64(7), "5VERv8NMvzbJMEkV8xnrLkEaWRtSz9CosKDYjCJjBRnbJ", nil, now, now))
		mock.ExpectQuery(`SELECT \* FROM "solana_latest_blocks" WHERE node_id.*LIMIT`).WithArgs(otherNodeID, sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows(columns).
				AddRow(uuid.New(), otherNodeID, uint64(123), "5VERv8NMvzbJMEkV8xnrLkEaWRtSz9CosKDYjCJjBRnbJ", nil, now, now))

		slot, err := s.Get(context.Background(), nodeID)
		require.NoError(t, err)
		assert.Equal(t, uint64(7), slot)

		otherSlot, err := s.Get(context.Background(), otherNodeID)
		require.NoError(t, err)
		assert.Equal(t, uint64(123), otherSlot)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("NotFoundReturnsZero", func(t *testing.T) {
		s, mock := newMockedLatestBlockStore(t)
		nodeID := uuid.New()
		mock.ExpectQuery(`SELECT \* FROM "solana_latest_blocks" WHERE node_id.*LIMIT`).WithArgs(nodeID, sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows(columns))

		slot, err := s.Get(context.Background(), nodeID)

		require.NoError(t, err)
		assert.Equal(t, uint64(0), slot)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("WrapsUnexpectedError", func(t *testing.T) {
		s, mock := newMockedLatestBlockStore(t)
		nodeID := uuid.New()
		dbErr := errors.New("connection reset")
		mock.ExpectQuery(`SELECT \* FROM "solana_latest_blocks" WHERE node_id.*LIMIT`).WithArgs(nodeID, sqlmock.AnyArg()).
			WillReturnError(dbErr)

		slot, err := s.Get(context.Background(), nodeID)

		require.Error(t, err)
		assert.ErrorIs(t, err, dbErr)
		assert.ErrorContains(t, err, "gorm solana latest block store:")
		assert.Equal(t, uint64(0), slot)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestNewSolanaLatestBlockStore(t *testing.T) {
	db := &gorm.DB{}

	s := NewSolanaLatestBlockStore(db)

	assert.Same(t, db, s.DB())
}
