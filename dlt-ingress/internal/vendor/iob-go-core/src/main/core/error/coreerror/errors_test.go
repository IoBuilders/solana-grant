package coreerror

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoreError_NotFoundError_WrapsSentinel(t *testing.T) {
	code := ErrorCode("ACTOR_NOT_FOUND")
	err := NewNotFoundDomainError(code, "actor not found")

	assert.True(t, errors.Is(err, ErrNotFound))
	assert.False(t, errors.Is(err, ErrValidation))

	var domErr DomainError
	assert.True(t, errors.As(err, &domErr))
	assert.Equal(t, code, domErr.ErrorCode())
	assert.Equal(t, "actor not found", domErr.Error())
}

func TestCoreError_ConflictError_WrapsSentinel(t *testing.T) {
	code := ErrorCode("ACTOR_ALREADY_EXISTS")
	err := NewConflictDomainError(code, "actor already exists")

	assert.True(t, errors.Is(err, ErrConflict))

	var domErr DomainError
	assert.True(t, errors.As(err, &domErr))
	assert.Equal(t, code, domErr.ErrorCode())
	assert.Equal(t, "actor already exists", domErr.Error())
}

func TestCoreError_ValidationError_WrapsSentinel(t *testing.T) {
	code := ErrorCode("FIELD_REQUIRED")
	err := NewValidationDomainError(code, "field is required")

	assert.True(t, errors.Is(err, ErrValidation))
	assert.False(t, errors.Is(err, ErrNotFound))

	var domErr DomainError
	assert.True(t, errors.As(err, &domErr))
	assert.Equal(t, code, domErr.ErrorCode())
	assert.Equal(t, "field is required", domErr.Error())
}

func TestCoreError_UnauthorizedError_WrapsSentinel(t *testing.T) {
	code := ErrorCode("NO_AUTHORIZED")
	err := NewUnauthorizedDomainError(code, "no auth")

	assert.True(t, errors.Is(err, ErrUnauthorized))

	var domErr DomainError
	assert.True(t, errors.As(err, &domErr))
	assert.Equal(t, code, domErr.ErrorCode())
	assert.Equal(t, "no auth", domErr.Error())
}

func TestCoreError_ForbiddenError_WrapsSentinel(t *testing.T) {
	code := ErrorCode("NO_PERMISSION")
	err := NewForbiddenDomainError(code, "no permission")

	assert.True(t, errors.Is(err, ErrForbidden))

	var domErr DomainError
	assert.True(t, errors.As(err, &domErr))
	assert.Equal(t, code, domErr.ErrorCode())
	assert.Equal(t, "no permission", domErr.Error())
}

