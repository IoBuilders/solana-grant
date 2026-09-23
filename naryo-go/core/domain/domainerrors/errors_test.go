//go:build test

package domainerrors

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestDomainErrors_Sentinels(t *testing.T) {
	t.Run("Validation", func(t *testing.T) {
		err := NewEmptyFieldError("Value", "Name")
		assert.ErrorIs(t, err, ErrValidation)
		assert.NotErrorIs(t, err, ErrNotFound)

		var domainErr DomainError
		assert.ErrorAs(t, err, &domainErr)
		assert.Equal(t, ErrorCodeEmptyField, domainErr.ErrorCode())
		assert.Equal(t, "Field Value of Name cannot be empty", err.Error())
	})

	t.Run("NotFound", func(t *testing.T) {
		id := uuid.New()
		err := NewEntityNotFoundError("Node", id)
		assert.ErrorIs(t, err, ErrNotFound)
		assert.NotErrorIs(t, err, ErrValidation)

		var domainErr DomainError
		assert.ErrorAs(t, err, &domainErr)
		assert.Equal(t, ErrorCodeEntityNotFound, domainErr.ErrorCode())
		assert.Equal(t, "Node with id "+id.String()+" not found", err.Error())
	})

	t.Run("Decoding", func(t *testing.T) {
		err := NewBufferUnderflowError(0, 8, 3)
		assert.ErrorIs(t, err, ErrDecoding)
		assert.NotErrorIs(t, err, ErrValidation)
		assert.NotErrorIs(t, err, ErrNotFound)

		var domainErr DomainError
		assert.ErrorAs(t, err, &domainErr)
		assert.Equal(t, ErrorCodeBufferUnderflow, domainErr.ErrorCode())
		assert.Equal(t, "buffer underflow at position 0: need 8 bytes, have 3", err.Error())
	})
}
