//go:build test

package parameter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func TestSolanaStringParameter_New(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		p, err := NewSolanaStringParameter(2, "hello")
		assert.NoError(t, err)
		assert.Equal(t, TypeString, p.Type())
		assert.Equal(t, "hello", p.Value())
	})

	t.Run("NegativePosition", func(t *testing.T) {
		_, err := NewSolanaStringParameter(-1, "hello")
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}
