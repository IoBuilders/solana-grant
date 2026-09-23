//go:build test

package solana

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
)

func TestNewOptionParameterDefinition(t *testing.T) {
	inner, err := NewUintParameterDefinition(0, 64)
	require.NoError(t, err)

	t.Run("Valid", func(t *testing.T) {
		def, err := NewOptionParameterDefinition(0, inner)
		assert.NoError(t, err)
		assert.Equal(t, parameter.TypeOption, def.Type())
		assert.Equal(t, 0, def.Position())
		assert.Equal(t, inner, def.Inner)
	})

	t.Run("NegativePosition", func(t *testing.T) {
		_, err := NewOptionParameterDefinition(-1, inner)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("NilInner", func(t *testing.T) {
		_, err := NewOptionParameterDefinition(0, nil)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}
