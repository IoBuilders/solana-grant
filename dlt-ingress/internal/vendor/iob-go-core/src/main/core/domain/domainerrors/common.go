package domainerrors

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/utils"
)

const (
	ErrorCodeValidation coreerror.ErrorCode = "VALIDATION_ERROR"
	ErrorCodeNotAllowed coreerror.ErrorCode = "NOT_ALLOWED"
)

func NewEmptyFieldError(fieldName string, entityName string) error {
	return coreerror.NewValidationDomainError(ErrorCodeValidation, fmt.Sprintf("field '%s' is required and cannot be null for entity %s", fieldName, entityName))
}

func NewFieldTooLongError(fieldName string, currentLength, maxLength int, entityName string) error {
	return coreerror.NewValidationDomainError(ErrorCodeValidation, fmt.Sprintf("field '%s' is too long: got %d characters, maximum allowed is %d for entity %s", fieldName, currentLength, maxLength, entityName))
}

func NewNotAllowedDomainError() error {
	return coreerror.NewForbiddenDomainError(ErrorCodeNotAllowed, "You don't have permission to perform this action")
}

func NewEntityNotFoundDomainError[T uuid.UUID | string](entity string, id T) error {
	var msg string
	switch v := any(id).(type) {
	case uuid.UUID:
		if v != uuid.Nil {
			msg = fmt.Sprintf("Entity %s with id %s not found", entity, v)
		} else {
			msg = fmt.Sprintf("Entity %s not found", entity)
		}
	case string:
		if v != "" {
			msg = fmt.Sprintf("Entity %s with %s not found", entity, v)
		} else {
			msg = fmt.Sprintf("Entity %s not found", entity)
		}
	default:
		msg = fmt.Sprintf("Entity %s with unknown identifier type not found", entity)
	}

	return coreerror.NewNotFoundDomainError(
		coreerror.ErrorCode(fmt.Sprintf("%s_NOT_FOUND", strings.ToUpper(utils.ToSnakeCase(entity)))),
		msg,
	)
}

func NewEntityUnexpectedStatusDomainError(entity string, id uuid.UUID, current string, expected string) error {
	var msg string
	if expected == "" {
		msg = fmt.Sprintf("Entity %s with id %s is in status %s", entity, id, current)
	} else {
		msg = fmt.Sprintf("Entity %s with id %s is in status %s instead of %s", entity, id, current, expected)
	}
	return coreerror.NewConflictDomainError(
		coreerror.ErrorCode(fmt.Sprintf("%s_UNEXPECTED_STATUS", strings.ToUpper(utils.ToSnakeCase(entity)))),
		msg,
	)
}

func NewEntityAlreadyExistsDomainError[T uuid.UUID | string](entity string, id T) error {
	var msg string
	switch v := any(id).(type) {
	case uuid.UUID:
		if v != uuid.Nil {
			msg = fmt.Sprintf("Entity %s with id %s already exists", entity, v)
		} else {
			msg = fmt.Sprintf("Entity %s already exists", entity)
		}
	case string:
		if v != "" {
			msg = fmt.Sprintf("Entity %s with %s already exists", entity, v)
		} else {
			msg = fmt.Sprintf("Entity %s already exists", entity)
		}
	default:
		msg = fmt.Sprintf("Entity %s with unknown identifier type already exists", entity)
	}

	return coreerror.NewConflictDomainError(
		coreerror.ErrorCode(fmt.Sprintf("%s_ALREADY_EXISTS", strings.ToUpper(utils.ToSnakeCase(entity)))),
		msg,
	)
}
