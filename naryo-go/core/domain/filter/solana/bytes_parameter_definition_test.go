//go:build test

package solana

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
)

func TestNewBytesParameterDefinition(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		def, err := NewBytesParameterDefinition(3)
		assert.NoError(t, err)
		assert.Equal(t, parameter.TypeBytes, def.Type())
		assert.Equal(t, 3, def.Position())
	})

	t.Run("NegativePosition", func(t *testing.T) {
		_, err := NewBytesParameterDefinition(-1)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}
