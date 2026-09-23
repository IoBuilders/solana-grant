//go:build test

package solana

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
)

func TestNewBytesFixedParameterDefinition(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		def, err := NewBytesFixedParameterDefinition(0, 32)
		assert.NoError(t, err)
		assert.Equal(t, parameter.TypeBytesFixed, def.Type())
		assert.Equal(t, 0, def.Position())
		assert.Equal(t, 32, def.ByteLength)
	})

	t.Run("NegativePosition", func(t *testing.T) {
		_, err := NewBytesFixedParameterDefinition(-1, 32)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("InvalidByteLength", func(t *testing.T) {
		_, err := NewBytesFixedParameterDefinition(0, 0)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}
