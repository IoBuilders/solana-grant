package feature

import (
	"strings"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

// Destination is the logical place a store writes to, and it is optional
// because not every backend can honour one. An HTTP store needs it.
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
