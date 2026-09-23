//go:build test

package solana

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
)

func TestNewStringParameterDefinition(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		def, err := NewStringParameterDefinition(2)
		assert.NoError(t, err)
		assert.Equal(t, parameter.TypeString, def.Type())
		assert.Equal(t, 2, def.Position())
	})

	t.Run("NegativePosition", func(t *testing.T) {
		_, err := NewStringParameterDefinition(-1)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}
