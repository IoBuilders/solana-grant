//go:build test

package parameter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func TestSolanaOptionParameter_New(t *testing.T) {
	inner, err := NewSolanaStringParameter(0, "value")
	assert.NoError(t, err)

	t.Run("Some", func(t *testing.T) {
		p, err := NewSolanaOptionParameter(0, inner)
		assert.NoError(t, err)
		assert.Equal(t, TypeOption, p.Type())
		assert.Equal(t, inner, p.Value())
	})

	t.Run("None", func(t *testing.T) {
		p, err := NewSolanaOptionParameter(0, nil)
		assert.NoError(t, err)
		assert.Nil(t, p.Value())
	})

	t.Run("NegativePosition", func(t *testing.T) {
		_, err := NewSolanaOptionParameter(-1, inner)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}
