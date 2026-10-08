package coreerror

import (
	"errors"
)

type ErrorCode string

var (
	ErrNotFound        = errors.New("not found")
	ErrValidation      = errors.New("validation error")
	ErrConflict        = errors.New("conflict")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrForbidden       = errors.New("forbidden")
	ErrTooManyRequests = errors.New("too many requests")
)

type DomainError interface {
	error
	Unwrap() error
	ErrorCode() ErrorCode
}

type RetryAfterError interface {
	error
	RetryAfterSeconds() int
}

type domainError struct {
	sentinel error
	code     ErrorCode
	message  string
}

func (d *domainError) Error() string        { return d.message }
func (d *domainError) Unwrap() error        { return d.sentinel }
func (d *domainError) ErrorCode() ErrorCode { return d.code }

func NewNotFoundDomainError(errorCode ErrorCode, msg string) error {
	return &domainError{sentinel: ErrNotFound, code: errorCode, message: msg}
}

func NewValidationDomainError(errorCode ErrorCode, msg string) error {
	return &domainError{sentinel: ErrValidation, code: errorCode, message: msg}
}

func NewConflictDomainError(errorCode ErrorCode, msg string) error {
	return &domainError{sentinel: ErrConflict, code: errorCode, message: msg}
}

func NewUnauthorizedDomainError(errorCode ErrorCode, msg string) error {
	return &domainError{sentinel: ErrUnauthorized, code: errorCode, message: msg}
}

func NewForbiddenDomainError(errorCode ErrorCode, msg string) error {
	return &domainError{sentinel: ErrForbidden, code: errorCode, message: msg}
}
