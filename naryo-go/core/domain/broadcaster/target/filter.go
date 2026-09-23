package target

import (
	"github.com/google/uuid"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

// FilterTarget forwards the contract events matched by one Filter.
type FilterTarget struct {
	destinations []Destination
	FilterID     uuid.UUID
}

func NewFilterTarget(destinations []Destination, filterID uuid.UUID) (*FilterTarget, error) {
	t := &FilterTarget{destinations: destinations, FilterID: filterID}
	if err := t.Validate(); err != nil {
		return nil, err
	}
	return t, nil
}

func (t FilterTarget) Type() Type {
	return TypeFilter
}

func (t FilterTarget) Destinations() []Destination {
	return t.destinations
}

func (t FilterTarget) Validate() error {
	if t.FilterID == uuid.Nil {
		return domainerrors.NewEmptyFieldError("FilterID", "FilterTarget")
	}
	return validateDestinations(t.destinations, "FilterTarget")
}

var _ Target = FilterTarget{}
