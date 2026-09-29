package domainerrors

import (
	"fmt"

	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
)

const (
	ErrorCodeEventConsumerNotFound      coreerror.ErrorCode = "EVENT_CONSUMER_NOT_FOUND"
	ErrorCodeEventConsumerInvalidStatus coreerror.ErrorCode = "EVENT_CONSUMER_INVALID_STATUS"
)

func NewEventConsumerNotFoundDomainError(id uuid.UUID) error {
	return coreerror.NewNotFoundDomainError(ErrorCodeEventConsumerNotFound, fmt.Sprintf("EventConsumer with id %s not found", id))
}

func NewEventConsumerInvalidStatusFailureDomainError(status string) error {
	return coreerror.NewValidationDomainError(ErrorCodeEventConsumerInvalidStatus, fmt.Sprintf("Invalid event consumer status %s", status))
}
