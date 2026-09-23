package store

import (
	"github.com/google/uuid"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/store/feature"
)

// ActiveConfiguration is a Configuration that does persist: it names the
// backend and carries one feature.Configuration per enabled feature.Type.
type ActiveConfiguration struct {
	nodeID    uuid.UUID
	storeType Type
	features  map[feature.Type]feature.Configuration
}

func NewActiveConfiguration(
	nodeID uuid.UUID,
	storeType Type,
	features map[feature.Type]feature.Configuration,
) (*ActiveConfiguration, error) {
	c := &ActiveConfiguration{nodeID: nodeID, storeType: storeType, features: features}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c ActiveConfiguration) NodeID() uuid.UUID {
	return c.nodeID
}

func (c ActiveConfiguration) State() State {
	return StateActive
}

func (c ActiveConfiguration) Type() Type {
	return c.storeType
}

func (c ActiveConfiguration) Feature(featureType feature.Type) (feature.Configuration, bool) {
	f, ok := c.features[featureType]
	return f, ok
}

func (c ActiveConfiguration) Validate() error {
	if c.nodeID == uuid.Nil {
		return domainerrors.NewEmptyFieldError("NodeID", "ActiveConfiguration")
	}
	if !c.storeType.IsValid() {
		return domainerrors.NewInvalidFieldError("Type", "ActiveConfiguration", "unknown store type "+c.storeType.String())
	}
	if len(c.features) == 0 {
		return domainerrors.NewEmptyFieldError("Features", "ActiveConfiguration")
	}
	for featureType, f := range c.features {
		if !featureType.IsValid() {
			return domainerrors.NewInvalidFieldError("Features", "ActiveConfiguration", "unknown feature type "+featureType.String())
		}
		if f == nil {
			return domainerrors.NewEmptyFieldError("Features["+featureType.String()+"]", "ActiveConfiguration")
		}
		if f.Type() != featureType {
			return domainerrors.NewInvalidFieldError(
				"Features", "ActiveConfiguration",
				"feature registered under "+featureType.String()+" reports type "+f.Type().String(),
			)
		}
		if err := f.Validate(); err != nil {
			return err
		}
	}
	return nil
}

var _ Configuration = (*ActiveConfiguration)(nil)
