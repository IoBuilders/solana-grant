package filter

import (
	"strings"

	"github.com/google/uuid"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

// FilterType discriminates the kind of a Filter: one that matches ingested
// events (EVENT) or one that matches transactions, chiefly failed ones
// (TRANSACTION).
type FilterType string

const (
	FilterTypeEvent       FilterType = "EVENT"
	FilterTypeTransaction FilterType = "TRANSACTION"
)

func (t FilterType) IsValid() bool {
	switch t {
	case FilterTypeEvent, FilterTypeTransaction:
		return true
	}
	return false
}

func (t FilterType) String() string {
	return string(t)
}

// Filter is the identity shared by every filter kind. Concrete filters
// (EventFilter, TransactionFilter) embed it, so its fields are promoted and
// its validation is reused across kinds.
type Filter interface {
	ID() uuid.UUID
	Name() Name
	NodeID() uuid.UUID
	Type() FilterType
}

type filter struct {
	id         uuid.UUID
	name       Name
	nodeID     uuid.UUID
	filterType FilterType
}

func (f *filter) ID() uuid.UUID {
	return f.id
}

func (f *filter) Name() Name {
	return f.name
}

func (f *filter) NodeID() uuid.UUID {
	return f.nodeID
}

func (f *filter) Type() FilterType {
	return f.filterType
}

func (f *filter) validate() error {
	if f.id == uuid.Nil {
		return domainerrors.NewEmptyFieldError("ID", "Filter")
	}
	if strings.TrimSpace(string(f.name)) == "" {
		return domainerrors.NewEmptyFieldError("Name", "Filter")
	}
	if f.nodeID == uuid.Nil {
		return domainerrors.NewEmptyFieldError("NodeID", "Filter")
	}
	if !f.filterType.IsValid() {
		return domainerrors.NewInvalidFieldError("Type", "Filter", "unsupported type "+f.filterType.String())
	}
	return nil
}
