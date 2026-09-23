//go:build test

package filter

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
)

func validSyncState(t *testing.T, filterID uuid.UUID) *SyncState {
	t.Helper()
	state, err := NewSyncState(filterID, 0, SyncStatusIdle)
	require.NoError(t, err)
	return state
}

// stubSpecification is a minimal Specification used to exercise the Filter
// aggregate without pulling in any chain-specific package.
type stubSpecification struct{}

func (stubSpecification) Strategy() Strategy               { return Strategy("STUB") }
func (stubSpecification) EventName() string                { return "stub-event" }
func (stubSpecification) Matches(event.ContractEvent) bool { return false }

func validSpec() Specification {
	return stubSpecification{}
}

func ptr(s string) *string {
	return &s
}

func TestNewEventFilter(t *testing.T) {
	id := uuid.New()
	nodeID := uuid.New()
	name, err := NewName("transfers")
	require.NoError(t, err)

	t.Run("ValidContractScope", func(t *testing.T) {
		f, err := NewEventFilter(id, name, nodeID, ScopeContract, ptr("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"), validSpec(), nil, validSyncState(t, id))
		assert.NoError(t, err)
		assert.NotNil(t, f)
		assert.Equal(t, ScopeContract, f.Scope)
	})

	t.Run("ValidGlobalScope", func(t *testing.T) {
		f, err := NewEventFilter(id, name, nodeID, ScopeGlobal, nil, validSpec(), nil, validSyncState(t, id))
		assert.NoError(t, err)
		assert.NotNil(t, f)
	})

	t.Run("ContractScopeWithoutContractID", func(t *testing.T) {
		_, err := NewEventFilter(id, name, nodeID, ScopeContract, nil, validSpec(), nil, validSyncState(t, id))
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("ContractScopeBlankContractID", func(t *testing.T) {
		_, err := NewEventFilter(id, name, nodeID, ScopeContract, ptr("   "), validSpec(), nil, validSyncState(t, id))
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("GlobalScopeWithContractID", func(t *testing.T) {
		_, err := NewEventFilter(id, name, nodeID, ScopeGlobal, ptr("SomeProgram"), validSpec(), nil, validSyncState(t, id))
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("NilID", func(t *testing.T) {
		// SyncState is built as a literal here: its constructor rejects a nil
		// FilterID, but the aggregate's ID check must fire first anyway.
		_, err := NewEventFilter(uuid.Nil, name, nodeID, ScopeGlobal, nil, validSpec(), nil, &SyncState{FilterID: uuid.Nil, Status: SyncStatusIdle})
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("BlankName", func(t *testing.T) {
		_, err := NewEventFilter(id, Name("  "), nodeID, ScopeGlobal, nil, validSpec(), nil, validSyncState(t, id))
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("NilNodeID", func(t *testing.T) {
		_, err := NewEventFilter(id, name, uuid.Nil, ScopeGlobal, nil, validSpec(), nil, validSyncState(t, id))
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("InvalidScope", func(t *testing.T) {
		_, err := NewEventFilter(id, name, nodeID, Scope("PROGRAM"), nil, validSpec(), nil, validSyncState(t, id))
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("NilSpecification", func(t *testing.T) {
		_, err := NewEventFilter(id, name, nodeID, ScopeGlobal, nil, nil, nil, validSyncState(t, id))
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("SyncStateOfAnotherFilter", func(t *testing.T) {
		_, err := NewEventFilter(id, name, nodeID, ScopeGlobal, nil, validSpec(), nil, validSyncState(t, uuid.New()))
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("DefaultsToAllStatuses", func(t *testing.T) {
		f, err := NewEventFilter(id, name, nodeID, ScopeGlobal, nil, validSpec(), nil, nil)
		require.NoError(t, err)
		assert.Equal(t, event.AllContractEventStatuses(), f.Statuses)
	})

	t.Run("InvalidStatus", func(t *testing.T) {
		_, err := NewEventFilter(id, name, nodeID, ScopeGlobal, nil, validSpec(), []event.ContractEventStatus{"BOGUS"}, nil)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}

// matchingSpecification is a Specification stub whose Matches result is
// configured per test.
type matchingSpecification struct {
	matches bool
}

func (matchingSpecification) Strategy() Strategy                 { return Strategy("STUB") }
func (matchingSpecification) EventName() string                  { return "stub-event" }
func (s matchingSpecification) Matches(event.ContractEvent) bool { return s.matches }

// stubContractEvent is a minimal ContractEvent test double.
type stubContractEvent struct {
	status event.ContractEventStatus
}

func (stubContractEvent) EventType() event.Type                          { return event.TypeContract }
func (stubContractEvent) NodeID() uuid.UUID                              { return uuid.Nil }
func (stubContractEvent) Parameters() []parameter.ContractEventParameter { return nil }
func (stubContractEvent) EventName() string                              { return "" }
func (e stubContractEvent) Status() event.ContractEventStatus            { return e.status }

func TestEventFilter_Matches(t *testing.T) {
	id := uuid.New()
	nodeID := uuid.New()
	name, err := NewName("transfers")
	require.NoError(t, err)

	t.Run("StatusNotAllowed", func(t *testing.T) {
		f, err := NewEventFilter(id, name, nodeID, ScopeGlobal, nil, matchingSpecification{matches: true},
			[]event.ContractEventStatus{event.ContractEventStatusFinalized}, nil)
		require.NoError(t, err)

		assert.False(t, f.Matches(stubContractEvent{status: event.ContractEventStatusProcessed}))
	})

	t.Run("StatusAllowedButSpecificationDoesNotMatch", func(t *testing.T) {
		f, err := NewEventFilter(id, name, nodeID, ScopeGlobal, nil, matchingSpecification{matches: false}, nil, nil)
		require.NoError(t, err)

		assert.False(t, f.Matches(stubContractEvent{status: event.ContractEventStatusConfirmed}))
	})

	t.Run("StatusAllowedAndSpecificationMatches", func(t *testing.T) {
		f, err := NewEventFilter(id, name, nodeID, ScopeGlobal, nil, matchingSpecification{matches: true}, nil, nil)
		require.NoError(t, err)

		assert.True(t, f.Matches(stubContractEvent{status: event.ContractEventStatusConfirmed}))
	})
}
