//go:build test

package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScope_IsValid(t *testing.T) {
	assert.True(t, ScopeContract.IsValid())
	assert.True(t, ScopeGlobal.IsValid())
	assert.False(t, Scope("PROGRAM").IsValid())
	assert.False(t, Scope("").IsValid())
}

func TestSyncStatus_IsValid(t *testing.T) {
	assert.True(t, SyncStatusActive.IsValid())
	assert.True(t, SyncStatusSyncing.IsValid())
	assert.True(t, SyncStatusIdle.IsValid())
	assert.False(t, SyncStatus("PAUSED").IsValid())
}

func TestFilterType_IsValid(t *testing.T) {
	assert.True(t, FilterTypeEvent.IsValid())
	assert.True(t, FilterTypeTransaction.IsValid())
	assert.False(t, FilterType("RAW").IsValid())
	assert.False(t, FilterType("").IsValid())
}

func TestIdentifierType_IsValid(t *testing.T) {
	assert.True(t, IdentifierTypeHash.IsValid())
	assert.True(t, IdentifierTypeToAddress.IsValid())
	assert.True(t, IdentifierTypeFromAddress.IsValid())
	assert.True(t, IdentifierTypeIdentityID.IsValid())
	assert.False(t, IdentifierType("BLOCK").IsValid())
	assert.False(t, IdentifierType("").IsValid())
}

func TestTransactionStatus_IsValid(t *testing.T) {
	assert.True(t, TransactionStatusFailed.IsValid())
	assert.True(t, TransactionStatusConfirmed.IsValid())
	assert.True(t, TransactionStatusUnconfirmed.IsValid())
	assert.False(t, TransactionStatus("PENDING").IsValid())
	assert.False(t, TransactionStatus("").IsValid())
}
