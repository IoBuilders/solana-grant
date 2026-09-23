package domainerrors

import "fmt"

const (
	ErrorCodeEmptyField      ErrorCode = "EMPTY_FIELD"
	ErrorCodeInvalidField    ErrorCode = "INVALID_FIELD"
	ErrorCodeEntityNotFound  ErrorCode = "ENTITY_NOT_FOUND"
	ErrorCodeBufferUnderflow ErrorCode = "BUFFER_UNDERFLOW"
)

// NewEmptyFieldError reports a required field that is missing or blank.
func NewEmptyFieldError(field, entity string) error {
	return NewValidationDomainError(
		ErrorCodeEmptyField,
		fmt.Sprintf("Field %s of %s cannot be empty", field, entity),
	)
}

// NewInvalidFieldError reports a field whose value violates a domain invariant.
func NewInvalidFieldError(field, entity, reason string) error {
	return NewValidationDomainError(
		ErrorCodeInvalidField,
		fmt.Sprintf("Field %s of %s is invalid: %s", field, entity, reason),
	)
}

// NewEntityNotFoundError reports a missing entity looked up by id.
func NewEntityNotFoundError(entity string, id fmt.Stringer) error {
	return NewNotFoundDomainError(
		ErrorCodeEntityNotFound,
		fmt.Sprintf("%s with id %s not found", entity, id),
	)
}

// NewBufferUnderflowError reports that a buffer had fewer bytes remaining at
// position than a decode step required.
func NewBufferUnderflowError(position, needed, available int) error {
	return NewDecodingDomainError(
		ErrorCodeBufferUnderflow,
		fmt.Sprintf("buffer underflow at position %d: need %d bytes, have %d", position, needed, available),
	)
}
