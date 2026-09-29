package validate

import (
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"
)

func StringMaxLength(value, fieldName, entityName string, maxLength int) error {
	if value == "" {
		return domainerrors.NewEmptyFieldError(fieldName, entityName)
	}
	if maxLength > 0 && len(value) > maxLength {
		return domainerrors.NewFieldTooLongError(fieldName, len(value), maxLength, entityName)
	}
	return nil
}

func StringNotEmpty(fieldName, entityName, value string) error {
	if value == "" {
		return domainerrors.NewEmptyFieldError(fieldName, entityName)
	}
	return nil
}
