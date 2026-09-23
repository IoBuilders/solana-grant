package domainerrors

import "errors"

// ErrorCode identifies a domain error in a stable, machine-readable way.
type ErrorCode string

// Sentinels wrapped by domain errors, so callers can match with
// errors.Is(err, ErrValidation) / errors.Is(err, ErrNotFound) / errors.Is(err, ErrDecoding).
var (
	ErrValidation = errors.New("validation error")
	ErrNotFound   = errors.New("not found error")
	ErrDecoding   = errors.New("decoding error")
)

// DomainError is implemented by every error produced by this package.
type DomainError interface {
	error
	Unwrap() error
	ErrorCode() ErrorCode
}

type domainError struct {
	sentinel error
	code     ErrorCode
	message  string
}

func (e *domainError) Error() string {
	return e.message
}

func (e *domainError) Unwrap() error {
	return e.sentinel
}

func (e *domainError) ErrorCode() ErrorCode {
	return e.code
}

// NewValidationDomainError builds a DomainError wrapping ErrValidation.
func NewValidationDomainError(code ErrorCode, message string) error {
	return &domainError{sentinel: ErrValidation, code: code, message: message}
}

// NewNotFoundDomainError builds a DomainError wrapping ErrNotFound.
func NewNotFoundDomainError(code ErrorCode, message string) error {
	return &domainError{sentinel: ErrNotFound, code: code, message: message}
}

// NewDecodingDomainError builds a DomainError wrapping ErrDecoding.
func NewDecodingDomainError(code ErrorCode, message string) error {
	return &domainError{sentinel: ErrDecoding, code: code, message: message}
}
