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

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

func newSolanaBlockEvent(t *testing.T, opts ...func(*solanaBlockEventFixture)) event.SolanaBlockEvent {
	t.Helper()

	blockTime := int64(1_700_000_000)
	f := solanaBlockEventFixture{
		nodeID:    uuid.New(),
		slot:      42,
		blockhash: "5VERv8NMvzbJMEkV8xnrLkEaWRtSz9CosKDYjCJjBRnbJ",
		blockTime: &blockTime,
	}
	for _, opt := range opts {
		opt(&f)
	}

	e, err := event.NewSolanaBlockEvent(f.nodeID, f.slot, f.blockhash, f.blockTime, f.transactions)
	require.NoError(t, err)

	return e
}

type solanaBlockEventFixture struct {
	nodeID       uuid.UUID
	slot         uint64
	blockhash    string
	blockTime    *int64
	transactions []event.SolanaTransaction
}

func withTransaction(nodeID uuid.UUID, slot uint64) func(*solanaBlockEventFixture) {
	return func(f *solanaBlockEventFixture) {
		tx, err := event.NewSolanaTransaction(
			nodeID,
			"5VERv8NMvzbJMEkV8xnrLkEaWRtSz9CosKDYjCJjBRnbJLgp8uirBgmQpjKhoR4tjF3ZpRzrFmBV6UjKdiSZkQUW",
			slot,
			[]event.InstructionData{
				{ProgramID: "11111111111111111111111111111111", Data: []byte{1, 2, 3}, Accounts: []string{"acct-1", "acct-2"}},
			},
			[]string{"tx-acct-1", "tx-acct-2"},
			[]string{"log line"},
			nil,
		)
		if err != nil {
			panic(err)
		}
		f.transactions = append(f.transactions, tx)
	}
}

func newMockedBlockStore(t *testing.T) (*SolanaBlockEventStore, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	require.NoError(t, err)

	mock.MatchExpectationsInOrder(false)

	return NewSolanaBlockEventStore(gormDB), mock
}

func TestSolanaBlockEventStore_Save(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		s, mock := newMockedBlockStore(t)
		mock.ExpectBegin()
		// Only the block row itself: no transactions to cascade into.
		mock.ExpectExec(`INSERT INTO "solana_block_events"`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := s.Save(context.Background(), newSolanaBlockEvent(t))

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("SuccessWithTransaction", func(t *testing.T) {
		s, mock := newMockedBlockStore(t)
		nodeID := uuid.New()
		mock.ExpectBegin()
		// One exec per table touched: the block itself, the associated block
		// transaction, that transaction's instructions, the instructions'
		// accounts, the transaction's own accounts, and the transaction's logs.
		// None of these must ever touch "solana_transaction_events" or its
		// children: that family is reserved for filter-matched transactions,
		// saved independently via SolanaTransactionEventStore.
		mock.ExpectExec(`INSERT INTO "solana_block_events"`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO "solana_block_transactions"`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO "solana_block_instructions"`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO "solana_block_instruction_accounts"`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO "solana_block_transaction_accounts"`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO "solana_block_transaction_logs"`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		blockEvent := newSolanaBlockEvent(t, func(f *solanaBlockEventFixture) {
			f.nodeID = nodeID
		}, withTransaction(nodeID, 42))

		err := s.Save(context.Background(), blockEvent)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("WrapsPersistenceFailure", func(t *testing.T) {
		s, mock := newMockedBlockStore(t)
		dbErr := errors.New("connection reset")
		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "solana_block_events"`).WillReturnError(dbErr)
		mock.ExpectRollback()

		err := s.Save(context.Background(), newSolanaBlockEvent(t))

		require.Error(t, err)
		assert.ErrorIs(t, err, dbErr)
		assert.ErrorContains(t, err, "gorm solana block event store:")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSolanaBlockEventStore_GetLatest(t *testing.T) {
	columns := []string{"id", "node_id", "slot", "blockhash", "block_time", "created_at", "updated_at"}

	t.Run("Success", func(t *testing.T) {
		s, mock := newMockedBlockStore(t)
		nodeID := uuid.New()
		now := time.Now()
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows(columns).
			AddRow(uuid.New(), nodeID, uint64(99), "5VERv8NMvzbJMEkV8xnrLkEaWRtSz9CosKDYjCJjBRnbJ", nil, now, now))

		slot, err := s.GetLatest(context.Background(), nodeID)

		require.NoError(t, err)
		assert.Equal(t, uint64(99), slot)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("NotFoundReturnsZero", func(t *testing.T) {
		s, mock := newMockedBlockStore(t)
		nodeID := uuid.New()
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows(columns))

		slot, err := s.GetLatest(context.Background(), nodeID)

		require.NoError(t, err)
		assert.Equal(t, uint64(0), slot)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("PropagatesUnexpectedError", func(t *testing.T) {
		s, mock := newMockedBlockStore(t)
		nodeID := uuid.New()
		dbErr := errors.New("connection reset")
		mock.ExpectQuery(".*").WillReturnError(dbErr)

		slot, err := s.GetLatest(context.Background(), nodeID)

		assert.ErrorIs(t, err, dbErr)
		assert.Equal(t, uint64(0), slot)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestNewBlockEventStore(t *testing.T) {
	db := &gorm.DB{}

	s := NewSolanaBlockEventStore(db)

	assert.Same(t, db, s.DB())
}
