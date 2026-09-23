package filter

import (
	"strings"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

// Name is the human-readable, non-blank identifier of a Filter.
type Name string

func NewName(value string) (Name, error) {
	if strings.TrimSpace(value) == "" {
		return "", domainerrors.NewEmptyFieldError("Value", "Name")
	}
	return Name(value), nil
}

func (n Name) String() string {
	return string(n)
}
