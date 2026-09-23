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
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
)

func newSolanaContractEvent(t *testing.T, opts ...func(*solanaContractEventFixture)) event.SolanaContractEvent {
	t.Helper()

	param, err := parameter.NewSolanaStringParameter(0, "value")
	require.NoError(t, err)

	f := solanaContractEventFixture{
		nodeID:     uuid.New(),
		programID:  "11111111111111111111111111111111",
		signature:  "5VERv8NMvzbJMEkV8xnrLkEaWRtSz9CosKDYjCJjBRnbJLgp8uirBgmQpjKhoR4tjF3ZpRzrFmBV6UjKdiSZkQUW",
		slot:       42,
		parameters: []parameter.ContractEventParameter{param},
		eventName:  "Transfer",
		status:     event.ContractEventStatusConfirmed,
	}
	for _, opt := range opts {
		opt(&f)
	}

	e, err := event.NewSolanaContractEvent(f.nodeID, f.programID, f.signature, f.slot, f.parameters, f.eventName, f.status)
	require.NoError(t, err)

	return e
}

type solanaContractEventFixture struct {
	nodeID     uuid.UUID
	programID  string
	signature  string
	slot       uint64
	parameters []parameter.ContractEventParameter
	eventName  string
	status     event.ContractEventStatus
}

func withoutParameters(f *solanaContractEventFixture) {
	f.parameters = nil
}

func newMockedContractStore(t *testing.T) (*SolanaContractEventStore, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	require.NoError(t, err)

	mock.MatchExpectationsInOrder(false)

	return NewSolanaContractEventStore(gormDB), mock
}

func TestSolanaContractEventStore_Save(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		s, mock := newMockedContractStore(t)
		mock.ExpectBegin()
		// One exec per table touched: the contract event itself and its
		// decoded parameters.
		mock.ExpectExec(`INSERT INTO "solana_contract_events"`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO "contract_event_parameters"`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := s.Save(context.Background(), newSolanaContractEvent(t))

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("SuccessWithoutParameters", func(t *testing.T) {
		s, mock := newMockedContractStore(t)
		mock.ExpectBegin()
		// No decoded parameters to write, so only the contract event row
		// itself is inserted.
		mock.ExpectExec(`INSERT INTO "solana_contract_events"`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := s.Save(context.Background(), newSolanaContractEvent(t, withoutParameters))

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("WrapsPersistenceFailure", func(t *testing.T) {
		s, mock := newMockedContractStore(t)
		dbErr := errors.New("connection reset")
		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "solana_contract_events"`).WillReturnError(dbErr)
		mock.ExpectRollback()

		err := s.Save(context.Background(), newSolanaContractEvent(t))

		require.Error(t, err)
		assert.ErrorIs(t, err, dbErr)
		assert.ErrorContains(t, err, "gorm solana contract event store:")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestNewSolanaContractEventStore(t *testing.T) {
	db := &gorm.DB{}

	s := NewSolanaContractEventStore(db)

	assert.Same(t, db, s.DB())
}
