package validate

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
)

func TestStringMaxLength_ValidValue_ReturnsNil(t *testing.T) {
	assert.NoError(t, StringMaxLength("valid", "name", "Entity", 10))
}

func TestStringMaxLength_NoMaxLengthCheck_WhenZero(t *testing.T) {
	assert.NoError(t, StringMaxLength("anything goes", "name", "Entity", 0))
}

func TestStringMaxLength_Empty_ReturnsValidationDomainError(t *testing.T) {
	err := StringMaxLength("", "email", "Entity", 100)

	assert.Error(t, err)
	assert.Equal(t, "field 'email' is required and cannot be null for entity Entity", err.Error())

	var domErr coreerror.DomainError
	assert.True(t, errors.As(err, &domErr))
	assert.Equal(t, domainerrors.ErrorCodeValidation, domErr.ErrorCode())
	assert.True(t, errors.Is(err, coreerror.ErrValidation))
}

func TestStringMaxLength_TooLong_ReturnsValidationDomainError(t *testing.T) {
	err := StringMaxLength("abcdef", "name", "Entity", 3)

	assert.Error(t, err)
	assert.Equal(t, "field 'name' is too long: got 6 characters, maximum allowed is 3 for entity Entity", err.Error())

	var domErr coreerror.DomainError
	assert.True(t, errors.As(err, &domErr))
	assert.Equal(t, domainerrors.ErrorCodeValidation, domErr.ErrorCode())
	assert.True(t, errors.Is(err, coreerror.ErrValidation))
}

func TestStringNotEmpty_NonEmpty_ReturnsNil(t *testing.T) {
	assert.NoError(t, StringNotEmpty("name", "Entity", "value"))
}

func TestStringNotEmpty_Empty_ReturnsValidationDomainError(t *testing.T) {
	err := StringNotEmpty("name", "Entity", "")

	assert.Error(t, err)
	assert.Equal(t, "field 'name' is required and cannot be null for entity Entity", err.Error())

	var domErr coreerror.DomainError
	assert.True(t, errors.As(err, &domErr))
	assert.Equal(t, domainerrors.ErrorCodeValidation, domErr.ErrorCode())
	assert.True(t, errors.Is(err, coreerror.ErrValidation))
}
