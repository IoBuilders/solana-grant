//go:build test

package parameter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func TestSolanaArrayParameter_New(t *testing.T) {
	elem, err := NewSolanaBoolParameter(0, true)
	assert.NoError(t, err)

	t.Run("Valid", func(t *testing.T) {
		p, err := NewSolanaArrayParameter(0, []ContractEventParameter{elem})
		assert.NoError(t, err)
		assert.Equal(t, TypeArray, p.Type())
		assert.Equal(t, []ContractEventParameter{elem}, p.Value())
	})

	t.Run("NilValue", func(t *testing.T) {
		_, err := NewSolanaArrayParameter(0, nil)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("EmptyButNonNilValueIsValid", func(t *testing.T) {
		p, err := NewSolanaArrayParameter(0, []ContractEventParameter{})
		assert.NoError(t, err)
		assert.Equal(t, TypeArray, p.Type())
		assert.Empty(t, p.Value())
	})
}
