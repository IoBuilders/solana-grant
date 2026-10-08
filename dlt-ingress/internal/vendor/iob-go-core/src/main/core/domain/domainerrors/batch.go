package domainerrors

import (
	"fmt"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
)

const (
	ErrorCodeBatchInvalidEnvelope  coreerror.ErrorCode = "BATCH_INVALID_ENVELOPE"
	ErrorCodeBatchValidationFailed coreerror.ErrorCode = "BATCH_VALIDATION_FAILED"
	ErrorCodeBatchAllItemsInvalid  coreerror.ErrorCode = "BATCH_ALL_ITEMS_INVALID"
)

func NewBatchInvalidEnvelopeDomainError(field, reason string) error {
	return coreerror.NewConflictDomainError(ErrorCodeBatchInvalidEnvelope, fmt.Sprintf("field '%s' %s", field, reason))
}

func NewBatchValidationFailedDomainError(details string) error {
	return coreerror.NewConflictDomainError(ErrorCodeBatchValidationFailed, withDetails("Batch rejected: FAIL_FAST validation found invalid items", details))
}

func NewBatchAllItemsInvalidDomainError(details string) error {
	return coreerror.NewConflictDomainError(ErrorCodeBatchAllItemsInvalid, withDetails("Batch rejected: every item is invalid", details))
}

func withDetails(message, details string) string {
	if details == "" {
		return message
	}
	return message + ": " + details
}
