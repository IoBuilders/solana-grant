//go:build test

package parameter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func TestSolanaBytesFixedParameter_New(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		p, err := NewSolanaBytesFixedParameter(0, []byte{1, 2, 3, 4}, 4)
		assert.NoError(t, err)
		assert.Equal(t, TypeBytesFixed, p.Type())
		assert.Equal(t, []byte{1, 2, 3, 4}, p.Value())
		assert.Equal(t, 4, p.ByteLength)
	})

	t.Run("LengthMismatch", func(t *testing.T) {
		_, err := NewSolanaBytesFixedParameter(0, []byte{1, 2, 3}, 4)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("ZeroByteLength", func(t *testing.T) {
		_, err := NewSolanaBytesFixedParameter(0, []byte{}, 0)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}
