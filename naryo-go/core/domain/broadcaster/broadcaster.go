package broadcaster

import (
	"github.com/google/uuid"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster/target"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

// Broadcaster is the aggregate root binding what events to forward (Target)
// to the Configuration used to deliver them.
type Broadcaster struct {
	ID              uuid.UUID
	Target          target.Target
	ConfigurationID uuid.UUID
}

// NewBroadcaster builds a Broadcaster.
func NewBroadcaster(id uuid.UUID, t target.Target, configurationID uuid.UUID) (*Broadcaster, error) {
	b := &Broadcaster{
		ID:              id,
		Target:          t,
		ConfigurationID: configurationID,
	}
	if err := b.validate(); err != nil {
		return nil, err
	}
	return b, nil
}

func (b *Broadcaster) validate() error {
	if b.ID == uuid.Nil {
		return domainerrors.NewEmptyFieldError("ID", "Broadcaster")
	}
	if b.Target == nil {
		return domainerrors.NewEmptyFieldError("Target", "Broadcaster")
	}
	if err := b.Target.Validate(); err != nil {
		return err
	}
	if b.ConfigurationID == uuid.Nil {
		return domainerrors.NewEmptyFieldError("ConfigurationID", "Broadcaster")
	}
	return nil
}
