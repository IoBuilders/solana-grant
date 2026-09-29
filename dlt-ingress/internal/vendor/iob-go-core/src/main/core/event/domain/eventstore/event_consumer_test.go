package eventstore

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
)

func TestCoreEventConsumer_NewEventConsumer_Success(t *testing.T) {
	consumerType := "type"
	status := Pending
	eventConsumer, err := NewEventConsumer(consumerType, string(status), false)

	assert.Nil(t, err)
	assert.NotNil(t, eventConsumer)
	assert.Equal(t, eventConsumer.Type, consumerType)
	assert.Equal(t, eventConsumer.Status, status)
}

func TestCoreEventConsumer_NewEventConsumer_Type_Empty_Error(t *testing.T) {
	eventConsumer, err := NewEventConsumer("", string(Pending), false)

	assert.Nil(t, eventConsumer)
	assert.NotNil(t, err)
	assertValidationDomainError(t, err, domainerrors.ErrorCodeValidation)
}

func TestCoreEventConsumer_NewEventConsumer_Type_TooLong_Error(t *testing.T) {
	eventConsumer, err := NewEventConsumer(strings.Repeat("1", 256), string(Pending), false)

	assert.Nil(t, eventConsumer)
	assert.NotNil(t, err)
	assertValidationDomainError(t, err, domainerrors.ErrorCodeValidation)
}

func TestCoreEventConsumer_NewEventConsumer_Invalid_Status_Error(t *testing.T) {
	eventConsumer, err := NewEventConsumer("type", "invalidStatus", false)

	assert.Nil(t, eventConsumer)
	assert.NotNil(t, err)
	assertValidationDomainError(t, err, domainerrors.ErrorCodeEventConsumerInvalidStatus)
}

func assertValidationDomainError(t *testing.T, err error, expectedCode coreerror.ErrorCode) {
	t.Helper()
	var domErr coreerror.DomainError
	assert.True(t, errors.As(err, &domErr))
	assert.Equal(t, expectedCode, domErr.ErrorCode())
	assert.True(t, errors.Is(err, coreerror.ErrValidation))
}
