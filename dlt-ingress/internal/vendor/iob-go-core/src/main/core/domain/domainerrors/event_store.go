package domainerrors

import (
	"fmt"

	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
)

const (
	ErrorCodeEventStoreNotFound               coreerror.ErrorCode = "EVENT_STORE_NOT_FOUND"
	ErrorCodeEventStoreEmptyPayload           coreerror.ErrorCode = "EVENT_STORE_EMPTY_PAYLOAD"
	ErrorCodeEventStoreInvalidPublicationType coreerror.ErrorCode = "EVENT_STORE_INVALID_PUBLICATION_TYPE"
)

func NewEventStoreNotFoundDomainError(id uuid.UUID) error {
	return coreerror.NewNotFoundDomainError(ErrorCodeEventStoreNotFound, fmt.Sprintf("EventStore with id %s not found", id))
}

func NewEventStoreEmptyPayloadDomainError() error {
	return coreerror.NewValidationDomainError(ErrorCodeEventStoreEmptyPayload, "Event payload must not be empty")
}

func NewEventStoreInvalidPublicationTypeDomainError(publicationType string) error {
	return coreerror.NewValidationDomainError(ErrorCodeEventStoreInvalidPublicationType, fmt.Sprintf("Invalid event store publicationType %s", publicationType))
}
