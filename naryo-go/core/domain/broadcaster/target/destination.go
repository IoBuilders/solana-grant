package target

import (
	"strings"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

// Destination is where a Target delivers matched events, e.g. a path
// appended to a Configuration's endpoint, or a topic/routing key.
type Destination string

func NewDestination(value string) (Destination, error) {
	if strings.TrimSpace(value) == "" {
		return "", domainerrors.NewEmptyFieldError("Value", "Destination")
	}
	return Destination(value), nil
}

func (d Destination) String() string {
	return string(d)
}
