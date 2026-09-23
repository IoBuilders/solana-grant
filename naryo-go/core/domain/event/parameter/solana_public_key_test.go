//go:build test

package parameter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func TestSolanaPublicKeyParameter_New(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		p, err := NewSolanaPublicKeyParameter(0, "9xQeWvG816bUx9EPjHmaT23yvVM2ZWbrrpZb9PusVFin")
		assert.NoError(t, err)
		assert.Equal(t, TypePublicKey, p.Type())
		assert.Equal(t, "9xQeWvG816bUx9EPjHmaT23yvVM2ZWbrrpZb9PusVFin", p.Value())
	})

	t.Run("NegativePosition", func(t *testing.T) {
		_, err := NewSolanaPublicKeyParameter(-1, "9xQeWvG816bUx9EPjHmaT23yvVM2ZWbrrpZb9PusVFin")
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}
