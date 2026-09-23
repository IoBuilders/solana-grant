//go:build test

package solana

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
)

func TestNewUintParameterDefinition(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		def, err := NewUintParameterDefinition(0, 64)
		assert.NoError(t, err)
		assert.Equal(t, parameter.TypeUint, def.Type())
		assert.Equal(t, 0, def.Position())
		assert.Equal(t, 64, def.BitSize)
	})

	t.Run("NegativePosition", func(t *testing.T) {
		_, err := NewUintParameterDefinition(-1, 64)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("InvalidBitSize", func(t *testing.T) {
		_, err := NewUintParameterDefinition(0, 256)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("ValidBitSizes", func(t *testing.T) {
		for _, bitSize := range []int{8, 16, 32, 64, 128} {
			_, err := NewUintParameterDefinition(0, bitSize)
			assert.NoError(t, err)
		}
	})
}
