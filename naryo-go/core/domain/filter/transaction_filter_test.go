//go:build test

package filter

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

func TestNewTransactionFilter(t *testing.T) {
	id := uuid.New()
	nodeID := uuid.New()
	name, err := NewName("failed-txs")
	require.NoError(t, err)

	t.Run("Valid", func(t *testing.T) {
		f, err := NewTransactionFilter(id, name, nodeID, IdentifierTypeFromAddress, []string{"0xabc"}, []TransactionStatus{TransactionStatusFailed})
		require.NoError(t, err)
		assert.Equal(t, FilterTypeTransaction, f.Type())
		assert.Equal(t, IdentifierTypeFromAddress, f.IdentifierType)
		assert.Equal(t, []TransactionStatus{TransactionStatusFailed}, f.Statuses)
	})

	t.Run("EmptyStatusesDefaultsToAll", func(t *testing.T) {
		f, err := NewTransactionFilter(id, name, nodeID, IdentifierTypeHash, []string{"0xdeadbeef"}, nil)
		require.NoError(t, err)
		assert.Equal(t, AllTransactionStatuses(), f.Statuses)
	})

	t.Run("StatusesAreCopied", func(t *testing.T) {
		input := []TransactionStatus{TransactionStatusFailed}
		f, err := NewTransactionFilter(id, name, nodeID, IdentifierTypeHash, []string{"0xabc"}, input)
		require.NoError(t, err)

		input[0] = TransactionStatusConfirmed // mutate caller's slice
		assert.Equal(t, TransactionStatusFailed, f.Statuses[0])
	})

	t.Run("ValueIsCopied", func(t *testing.T) {
		input := []string{"0xabc"}
		f, err := NewTransactionFilter(id, name, nodeID, IdentifierTypeHash, input, nil)
		require.NoError(t, err)

		input[0] = "0xdef" // mutate caller's slice
		assert.Equal(t, "0xabc", f.Value[0])
	})

	t.Run("InvalidIdentifierType", func(t *testing.T) {
		_, err := NewTransactionFilter(id, name, nodeID, IdentifierType("BLOCK"), []string{"0xabc"}, nil)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("EmptyValue", func(t *testing.T) {
		_, err := NewTransactionFilter(id, name, nodeID, IdentifierTypeHash, nil, nil)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("BlankValueEntry", func(t *testing.T) {
		_, err := NewTransactionFilter(id, name, nodeID, IdentifierTypeHash, []string{"   "}, nil)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("UnknownStatus", func(t *testing.T) {
		_, err := NewTransactionFilter(id, name, nodeID, IdentifierTypeHash, []string{"0xabc"}, []TransactionStatus{"PENDING"})
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("NilID", func(t *testing.T) {
		_, err := NewTransactionFilter(uuid.Nil, name, nodeID, IdentifierTypeHash, []string{"0xabc"}, nil)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}

func TestTransactionFilter_Matches(t *testing.T) {
	id := uuid.New()
	nodeID := uuid.New()
	name, err := NewName("test-filter")
	require.NoError(t, err)

	tx, err := event.NewSolanaTransaction(nodeID, "sig-123", 42, nil, []string{"acc-1", "acc-2", "acc-3"}, nil, nil)
	require.NoError(t, err)

	t.Run("Hash_Matches", func(t *testing.T) {
		f, err := NewTransactionFilter(id, name, nodeID, IdentifierTypeHash, []string{"sig-123"}, nil)
		require.NoError(t, err)
		assert.True(t, f.Matches(tx))
	})

	t.Run("Hash_MatchesAnyEntry", func(t *testing.T) {
		f, err := NewTransactionFilter(id, name, nodeID, IdentifierTypeHash, []string{"sig-456", "sig-123"}, nil)
		require.NoError(t, err)
		assert.True(t, f.Matches(tx))
	})

	t.Run("Hash_NoMatch", func(t *testing.T) {
		f, err := NewTransactionFilter(id, name, nodeID, IdentifierTypeHash, []string{"sig-456"}, nil)
		require.NoError(t, err)
		assert.False(t, f.Matches(tx))
	})

	t.Run("ToAddress_NeverMatches", func(t *testing.T) {
		f, err := NewTransactionFilter(id, name, nodeID, IdentifierTypeToAddress, []string{"sig-123"}, nil)
		require.NoError(t, err)
		assert.False(t, f.Matches(tx))
	})

	t.Run("FromAddress_NeverMatches", func(t *testing.T) {
		f, err := NewTransactionFilter(id, name, nodeID, IdentifierTypeFromAddress, []string{"sig-123"}, nil)
		require.NoError(t, err)
		assert.False(t, f.Matches(tx))
	})

	t.Run("IdentityID_NeverMatches", func(t *testing.T) {
		f, err := NewTransactionFilter(id, name, nodeID, IdentifierTypeIdentityID, []string{"sig-123"}, nil)
		require.NoError(t, err)
		assert.False(t, f.Matches(tx))
	})

	t.Run("Addresses_MatchesWhenAllPresent", func(t *testing.T) {
		f, err := NewTransactionFilter(id, name, nodeID, IdentifierTypeAddresses, []string{"acc-1", "acc-3"}, nil)
		require.NoError(t, err)
		assert.True(t, f.Matches(tx))
	})

	t.Run("Addresses_NoMatchWhenOneMissing", func(t *testing.T) {
		f, err := NewTransactionFilter(id, name, nodeID, IdentifierTypeAddresses, []string{"acc-1", "acc-missing"}, nil)
		require.NoError(t, err)
		assert.False(t, f.Matches(tx))
	})

	t.Run("Addresses_NonSolanaTransaction_NeverMatches", func(t *testing.T) {
		f, err := NewTransactionFilter(id, name, nodeID, IdentifierTypeAddresses, []string{"acc-1"}, nil)
		require.NoError(t, err)
		assert.False(t, f.Matches(stubBlockTransaction{}))
	})

	t.Run("Status_ConfirmedMatchesWhenFiltered", func(t *testing.T) {
		f, err := NewTransactionFilter(id, name, nodeID, IdentifierTypeHash, []string{"sig-123"}, []TransactionStatus{TransactionStatusConfirmed})
		require.NoError(t, err)
		assert.True(t, f.Matches(tx))
	})

	t.Run("Status_FailedDoesNotMatchConfirmedTx", func(t *testing.T) {
		f, err := NewTransactionFilter(id, name, nodeID, IdentifierTypeHash, []string{"sig-123"}, []TransactionStatus{TransactionStatusFailed})
		require.NoError(t, err)
		assert.False(t, f.Matches(tx))
	})

	t.Run("Status_UnconfirmedNeverMatches", func(t *testing.T) {
		errMsg := "boom"
		failedTx, err := event.NewSolanaTransactionEvent(nodeID, "sig-456", 43, nil, []string{"acc-1"}, nil, &errMsg)
		require.NoError(t, err)

		f, err := NewTransactionFilter(id, name, nodeID, IdentifierTypeHash, []string{"sig-123", "sig-456"}, []TransactionStatus{TransactionStatusUnconfirmed})
		require.NoError(t, err)
		assert.False(t, f.Matches(tx), "confirmed tx should not match an Unconfirmed-only filter")
		assert.False(t, f.Matches(failedTx), "failed tx should not match an Unconfirmed-only filter")
	})
}

// stubBlockTransaction is a minimal BlockTransaction that is deliberately
// not a SolanaTransaction.
type stubBlockTransaction struct{}

func (stubBlockTransaction) Id() string     { return "sig-123" }
func (stubBlockTransaction) HasError() bool { return false }
