//go:build test

package solana

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
)

func TestNewArrayParameterDefinition(t *testing.T) {
	elem, err := NewUintParameterDefinition(0, 64)
	require.NoError(t, err)

	t.Run("ValidDynamic", func(t *testing.T) {
		def, err := NewArrayParameterDefinition(0, elem, nil)
		assert.NoError(t, err)
		assert.Equal(t, parameter.TypeArray, def.Type())
		assert.Equal(t, 0, def.Position())
		assert.Equal(t, elem, def.Element)
		assert.Nil(t, def.Length)
	})

	t.Run("ValidFixed", func(t *testing.T) {
		length := 4
		def, err := NewArrayParameterDefinition(0, elem, &length)
		assert.NoError(t, err)
		require.NotNil(t, def.Length)
		assert.Equal(t, 4, *def.Length)
	})

	t.Run("NegativePosition", func(t *testing.T) {
		_, err := NewArrayParameterDefinition(-1, elem, nil)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("NilElement", func(t *testing.T) {
		_, err := NewArrayParameterDefinition(0, nil, nil)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("ZeroLength", func(t *testing.T) {
		length := 0
		_, err := NewArrayParameterDefinition(0, elem, &length)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}
