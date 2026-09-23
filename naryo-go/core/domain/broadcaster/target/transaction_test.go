//go:build test

package target

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func TestTransactionTarget_New(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		destinations := newDestinations(t, "/webhook")
		target, err := NewTransactionTarget(destinations)
		assert.NoError(t, err)
		assert.Equal(t, TypeTransaction, target.Type())
		assert.Equal(t, destinations, target.Destinations())
	})

	t.Run("EmptyDestinations", func(t *testing.T) {
		_, err := NewTransactionTarget(nil)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}
