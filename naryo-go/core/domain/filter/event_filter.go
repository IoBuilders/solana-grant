package filter

import (
	"strings"

	"github.com/google/uuid"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

// EventFilter matches a Node's ingested events: its identity and name, the
// scope and contract it targets, the chain-specific Specification used to
// match, the ContractEventStatus values it accepts, and its synchronization
// progress. ContractAddress is nil for GLOBAL scope, where no single
// contract is targeted.
type EventFilter struct {
	filter
	Scope           Scope
	ContractAddress *string
	Specification   Specification
	Statuses        []event.ContractEventStatus
	SyncState       *SyncState
}

func NewEventFilter(
	id uuid.UUID,
	name Name,
	nodeID uuid.UUID,
	scope Scope,
	contractAddress *string,
	specification Specification,
	statuses []event.ContractEventStatus,
	syncState *SyncState,
) (*EventFilter, error) {
	if len(statuses) == 0 {
		statuses = event.AllContractEventStatuses()
	}
	normalized := make([]event.ContractEventStatus, len(statuses))
	copy(normalized, statuses)

	f := &EventFilter{
		filter:          filter{id: id, name: name, nodeID: nodeID, filterType: FilterTypeEvent},
		Scope:           scope,
		ContractAddress: contractAddress,
		Specification:   specification,
		Statuses:        normalized,
		SyncState:       syncState,
	}
	if err := f.Validate(); err != nil {
		return nil, err
	}
	return f, nil
}

func (f *EventFilter) Validate() error {
	if err := f.filter.validate(); err != nil {
		return err
	}
	if !f.Scope.IsValid() {
		return domainerrors.NewInvalidFieldError("Scope", "EventFilter", "unsupported scope "+f.Scope.String())
	}
	if f.Specification == nil {
		return domainerrors.NewEmptyFieldError("Specification", "EventFilter")
	}
	if err := f.validateScope(); err != nil {
		return err
	}
	if len(f.Statuses) == 0 {
		return domainerrors.NewEmptyFieldError("Statuses", "EventFilter")
	}
	for _, status := range f.Statuses {
		if !status.IsValid() {
			return domainerrors.NewInvalidFieldError("Statuses", "EventFilter", "unsupported status "+status.String())
		}
	}
	if f.SyncState != nil {
		return f.validateSyncState()
	}
	return nil
}

// Matches reports whether e satisfies this filter: its Status is one of
// Statuses, and its Specification matches e.
func (f *EventFilter) Matches(e event.ContractEvent) bool {
	if !containsStatus(f.Statuses, e.Status()) {
		return false
	}
	return f.Specification.Matches(e)
}

func containsStatus(statuses []event.ContractEventStatus, s event.ContractEventStatus) bool {
	for _, status := range statuses {
		if status == s {
			return true
		}
	}
	return false
}

func (f *EventFilter) validateScope() error {
	hasContractAddress := f.ContractAddress != nil && strings.TrimSpace(*f.ContractAddress) != ""
	switch f.Scope {
	case ScopeContract:
		if !hasContractAddress {
			return domainerrors.NewEmptyFieldError("ContractAddress", "EventFilter")
		}
	case ScopeGlobal:
		if hasContractAddress {
			return domainerrors.NewInvalidFieldError("ContractAddress", "EventFilter", "must be empty for GLOBAL scope")
		}
	}
	return nil
}

func (f *EventFilter) validateSyncState() error {
	if f.SyncState.FilterID != f.id {
		return domainerrors.NewInvalidFieldError("SyncState", "EventFilter", "sync state does not belong to this filter")
	}
	if !f.SyncState.Status.IsValid() {
		return domainerrors.NewInvalidFieldError("SyncState", "EventFilter", "unsupported status "+f.SyncState.Status.String())
	}
	return nil
}
