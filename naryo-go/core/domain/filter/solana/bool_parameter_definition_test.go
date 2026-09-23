//go:build test

package solana

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
)

func TestNewBoolParameterDefinition(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		def, err := NewBoolParameterDefinition(1)
		assert.NoError(t, err)
		assert.Equal(t, parameter.TypeBool, def.Type())
		assert.Equal(t, 1, def.Position())
	})

	t.Run("NegativePosition", func(t *testing.T) {
		_, err := NewBoolParameterDefinition(-1)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}
