//go:build test

package target

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func newDestinations(t *testing.T, values ...string) []Destination {
	t.Helper()
	destinations := make([]Destination, 0, len(values))
	for _, v := range values {
		d, err := NewDestination(v)
		assert.NoError(t, err)
		destinations = append(destinations, d)
	}
	return destinations
}

func TestDestination_New(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		d, err := NewDestination("/webhook")
		assert.NoError(t, err)
		assert.Equal(t, "/webhook", d.String())
	})

	t.Run("Zero", func(t *testing.T) {
		_, err := NewDestination("")
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("Blank", func(t *testing.T) {
		_, err := NewDestination("   ")
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}
