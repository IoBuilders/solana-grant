//go:build test

package eventstore

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

func newSolanaTransactionEvent(t *testing.T, opts ...func(*solanaTransactionEventFixture)) event.SolanaTransactionEvent {
	t.Helper()

	f := solanaTransactionEventFixture{
		nodeID:    uuid.New(),
		signature: "5VERv8NMvzbJMEkV8xnrLkEaWRtSz9CosKDYjCJjBRnbJLgp8uirBgmQpjKhoR4tjF3ZpRzrFmBV6UjKdiSZkQUW",
		slot:      42,
		instructions: []event.InstructionData{
			{ProgramID: "11111111111111111111111111111111", Data: []byte{1, 2, 3}, Accounts: []string{"acct-1", "acct-2"}},
		},
		accounts: []string{"tx-acct-1", "tx-acct-2"},
		logs:     []string{"log line"},
	}
	for _, opt := range opts {
		opt(&f)
	}

	e, err := event.NewSolanaTransactionEvent(f.nodeID, f.signature, f.slot, f.instructions, f.accounts, f.logs, f.err)
	require.NoError(t, err)

	return e
}

type solanaTransactionEventFixture struct {
	nodeID       uuid.UUID
	signature    string
	slot         uint64
	instructions []event.InstructionData
	accounts     []string
	logs         []string
	err          *string
}

func withoutInstructionsAccountsOrLogs(f *solanaTransactionEventFixture) {
	f.instructions = nil
	f.accounts = nil
	f.logs = nil
}

func newMockedStore(t *testing.T) (*SolanaTransactionEventStore, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	require.NoError(t, err)

	mock.MatchExpectationsInOrder(false)

	return NewSolanaTransactionEventStore(gormDB), mock
}

func TestSolanaTransactionEventStore_Save(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		s, mock := newMockedStore(t)
		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "solana_transaction_events"`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO "solana_instructions"`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO "solana_instruction_accounts"`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO "solana_transaction_accounts"`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO "solana_transaction_logs"`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := s.Save(context.Background(), newSolanaTransactionEvent(t))

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("SuccessWithoutInstructionsAccountsOrLogs", func(t *testing.T) {
		s, mock := newMockedStore(t)
		mock.ExpectBegin()
		// No instructions, accounts, or logs to write, so only the transaction
		// event row itself is inserted.
		mock.ExpectExec(`INSERT INTO "solana_transaction_events"`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := s.Save(context.Background(), newSolanaTransactionEvent(t, withoutInstructionsAccountsOrLogs))

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("WrapsPersistenceFailure", func(t *testing.T) {
		s, mock := newMockedStore(t)
		dbErr := errors.New("connection reset")
		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "solana_transaction_events"`).WillReturnError(dbErr)
		mock.ExpectRollback()

		err := s.Save(context.Background(), newSolanaTransactionEvent(t))

		require.Error(t, err)
		assert.ErrorIs(t, err, dbErr)
		assert.ErrorContains(t, err, "gorm solana transaction event store:")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestNewTransactionEventStore(t *testing.T) {
	db := &gorm.DB{}

	s := NewSolanaTransactionEventStore(db)

	assert.Same(t, db, s.DB())
}
