//go:build test

package parameter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func TestSolanaStructParameter_New(t *testing.T) {
	field, err := NewSolanaStringParameter(0, "value")
	assert.NoError(t, err)

	t.Run("Valid", func(t *testing.T) {
		p, err := NewSolanaStructParameter(0, []ContractEventParameter{field})
		assert.NoError(t, err)
		assert.Equal(t, TypeStruct, p.Type())
		assert.Equal(t, []ContractEventParameter{field}, p.Value())
	})

	t.Run("EmptyValue", func(t *testing.T) {
		_, err := NewSolanaStructParameter(0, nil)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}
