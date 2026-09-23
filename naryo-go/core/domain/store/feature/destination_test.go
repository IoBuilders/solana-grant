//go:build test

package feature

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func TestDestination_New(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		d, err := NewDestination("transactions")
		assert.NoError(t, err)
		assert.Equal(t, "transactions", d.String())
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
