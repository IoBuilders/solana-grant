//go:build test

package solana

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
)

func TestNewIntParameterDefinition(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		def, err := NewIntParameterDefinition(0, 64)
		assert.NoError(t, err)
		assert.Equal(t, parameter.TypeInt, def.Type())
		assert.Equal(t, 0, def.Position())
		assert.Equal(t, 64, def.BitSize)
	})

	t.Run("NegativePosition", func(t *testing.T) {
		_, err := NewIntParameterDefinition(-1, 64)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("InvalidBitSize", func(t *testing.T) {
		_, err := NewIntParameterDefinition(0, 256)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}
