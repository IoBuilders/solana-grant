package target

import (
	"strings"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

// Target describes which events a Broadcaster forwards and where.
type Target interface {
	Type() Type
	Destinations() []Destination
	Validate() error
}

func validateDestinations(destinations []Destination, entity string) error {
	if len(destinations) == 0 {
		return domainerrors.NewEmptyFieldError("Destinations", entity)
	}
	for _, d := range destinations {
		if strings.TrimSpace(string(d)) == "" {
			return domainerrors.NewInvalidFieldError("Destinations", entity, "destination cannot be blank")
		}
	}
	return nil
}
