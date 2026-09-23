//go:build test

package parameter

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func TestSolanaUintParameter_New(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		p, err := NewSolanaUintParameter(1, big.NewInt(42), 64)
		assert.NoError(t, err)
		assert.Equal(t, TypeUint, p.Type())
		assert.Equal(t, big.NewInt(42), p.Value())
		assert.Equal(t, 64, p.BitSize)
	})

	t.Run("NilValue", func(t *testing.T) {
		_, err := NewSolanaUintParameter(1, nil, 64)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("NegativeValue", func(t *testing.T) {
		_, err := NewSolanaUintParameter(1, big.NewInt(-1), 64)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("InvalidBitSize", func(t *testing.T) {
		_, err := NewSolanaUintParameter(1, big.NewInt(42), 24)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}
