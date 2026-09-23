//go:build test

package solana

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
)

func TestNewStructParameterDefinition(t *testing.T) {
	field0, err := NewUintParameterDefinition(0, 64)
	require.NoError(t, err)
	field1, err := NewBoolParameterDefinition(1)
	require.NoError(t, err)

	t.Run("Valid", func(t *testing.T) {
		def, err := NewStructParameterDefinition(0, []ParameterDefinition{field0, field1})
		assert.NoError(t, err)
		assert.Equal(t, parameter.TypeStruct, def.Type())
		assert.Equal(t, 0, def.Position())
		assert.Equal(t, []ParameterDefinition{field0, field1}, def.Fields)
	})

	t.Run("NegativePosition", func(t *testing.T) {
		_, err := NewStructParameterDefinition(-1, []ParameterDefinition{field0})
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("EmptyFields", func(t *testing.T) {
		_, err := NewStructParameterDefinition(0, nil)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("DuplicateFieldPosition", func(t *testing.T) {
		dupField, err := NewBoolParameterDefinition(0)
		require.NoError(t, err)
		_, err = NewStructParameterDefinition(0, []ParameterDefinition{field0, dupField})
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}
