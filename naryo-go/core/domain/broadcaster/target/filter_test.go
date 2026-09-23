//go:build test

package target

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func TestFilterTarget_New(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		destinations := newDestinations(t, "/webhook")
		filterID := uuid.New()

		target, err := NewFilterTarget(destinations, filterID)
		assert.NoError(t, err)
		assert.Equal(t, TypeFilter, target.Type())
		assert.Equal(t, destinations, target.Destinations())
		assert.Equal(t, filterID, target.FilterID)
	})

	t.Run("EmptyDestinations", func(t *testing.T) {
		_, err := NewFilterTarget(nil, uuid.New())
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("NilFilterID", func(t *testing.T) {
		_, err := NewFilterTarget(newDestinations(t, "/webhook"), uuid.Nil)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("BlankDestination", func(t *testing.T) {
		_, err := NewFilterTarget([]Destination{Destination("   ")}, uuid.New())
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}
