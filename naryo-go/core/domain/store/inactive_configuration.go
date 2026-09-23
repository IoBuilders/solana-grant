package store

import (
	"github.com/google/uuid"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

// InactiveConfiguration persists nothing for its Node.
type InactiveConfiguration struct {
	nodeID uuid.UUID
}

func NewInactiveConfiguration(nodeID uuid.UUID) (*InactiveConfiguration, error) {
	c := &InactiveConfiguration{nodeID: nodeID}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c InactiveConfiguration) NodeID() uuid.UUID {
	return c.nodeID
}

func (c InactiveConfiguration) State() State {
	return StateInactive
}

func (c InactiveConfiguration) Validate() error {
	if c.nodeID == uuid.Nil {
		return domainerrors.NewEmptyFieldError("NodeID", "InactiveConfiguration")
	}
	return nil
}

var _ Configuration = (*InactiveConfiguration)(nil)
