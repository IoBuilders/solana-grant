//go:build test

package node

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func TestName_New(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		name, err := NewName("solana-mainnet")
		assert.NoError(t, err)
		assert.Equal(t, Name("solana-mainnet"), name)
		assert.Equal(t, "solana-mainnet", name.String())
	})

	t.Run("Empty", func(t *testing.T) {
		_, err := NewName("")
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("Blank", func(t *testing.T) {
		_, err := NewName("   ")
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}
