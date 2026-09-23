//go:build test

package parameter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func TestSolanaBoolParameter_New(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		p, err := NewSolanaBoolParameter(0, true)
		assert.NoError(t, err)
		assert.Equal(t, TypeBool, p.Type())
		assert.Equal(t, 0, p.Position())
		assert.Equal(t, true, p.Value())
	})

	t.Run("NegativePosition", func(t *testing.T) {
		_, err := NewSolanaBoolParameter(-1, true)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}
