package broadcaster

import (
	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

// Configuration describes the transport a Broadcaster delivers events
// through. Its shape varies per Type, defined in the adapter module that owns it
type Configuration interface {
	ID() uuid.UUID
	Type() Type
	Validate() error
	AdditionalProperties() map[string]interface{}
}

// GenericConfiguration is a minimal Configuration for a Type whose real
// connection details live outside core, in the adapter module that owns it
type GenericConfiguration struct {
	id                   uuid.UUID
	typ                  Type
	additionalProperties map[string]interface{}
}

func NewGenericConfiguration(id uuid.UUID, typ Type, additionalProperties map[string]interface{}) (*GenericConfiguration, error) {
	c := &GenericConfiguration{id: id, typ: typ, additionalProperties: additionalProperties}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c GenericConfiguration) ID() uuid.UUID {
	return c.id
}

func (c GenericConfiguration) Type() Type {
	return c.typ
}

func (c GenericConfiguration) Validate() error {
	if c.id == uuid.Nil {
		return domainerrors.NewEmptyFieldError("ID", "GenericConfiguration")
	}
	if c.typ == "" {
		return domainerrors.NewEmptyFieldError("Type", "GenericConfiguration")
	}
	return nil
}

func (c GenericConfiguration) AdditionalProperties() map[string]interface{} {
	return c.additionalProperties
}

var _ Configuration = (*GenericConfiguration)(nil)
