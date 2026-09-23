//go:build test

package parameter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func TestSolanaFloatParameter_New(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		p, err := NewSolanaFloatParameter(0, 3.14, 64)
		assert.NoError(t, err)
		assert.Equal(t, TypeFloat, p.Type())
		assert.Equal(t, 3.14, p.Value())
		assert.Equal(t, 64, p.BitSize)
	})

	t.Run("InvalidBitSize", func(t *testing.T) {
		_, err := NewSolanaFloatParameter(0, 3.14, 16)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}
