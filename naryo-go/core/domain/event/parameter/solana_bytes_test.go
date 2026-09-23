//go:build test

package parameter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func TestSolanaBytesParameter_New(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		p, err := NewSolanaBytesParameter(0, []byte{1, 2, 3})
		assert.NoError(t, err)
		assert.Equal(t, TypeBytes, p.Type())
		assert.Equal(t, []byte{1, 2, 3}, p.Value())
	})

	t.Run("NilValue", func(t *testing.T) {
		_, err := NewSolanaBytesParameter(0, nil)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}
