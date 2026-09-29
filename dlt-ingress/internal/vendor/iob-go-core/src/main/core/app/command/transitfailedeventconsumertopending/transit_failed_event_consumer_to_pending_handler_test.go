package transitfailedeventconsumertopending

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/basemodel"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/domain/eventstore"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/domain/eventstore/events"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/port/repository/eventstore"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/test/core/event"
)

func buildventConsumer(status eventstore.Status) *eventstore.EventConsumer {
	return &eventstore.EventConsumer{
		Model:  basemodel.Model{Id: uuid.New()},
		Status: status,
	}
}

func TestCoreTransitFailedEventConsumerToPendingHandler_Handle_Success(t *testing.T) {
	eventBus := new(testevent.EventBusMock)
	eventConsumerRepo := new(eventstorerepo.EventConsumerRepositoryMock)
	handler := NewHandler(eventBus, eventConsumerRepo)

	eventConsumer := buildventConsumer(eventstore.Failed)
	eventConsumerRepo.On("FindById", mock.Anything, eventConsumer.Id).Return(eventConsumer, nil)
	eventConsumerRepo.On("Save", mock.Anything, mock.Anything).Return(nil)
	eventBus.On("Publish", mock.Anything, mock.MatchedBy(func(evt eventstoreevents.TransitFailedEventConsumerToPendingEvent) bool {
		return evt.EventConsumerId == eventConsumer.Id
	})).Return(nil).Once()

	cmd := Command{EventConsumerId: eventConsumer.Id}

	resp, err := handler.Handle(context.Background(), &cmd)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, eventConsumer.Id, resp.EventConsumerId)
	assert.Equal(t, eventstore.Pending, eventConsumer.Status)
	eventConsumerRepo.AssertExpectations(t)
}

func TestCoreTransitFailedEventConsumerToPendingHandler_Handle_NotificationNotFoundError(t *testing.T) {
	eventBus := new(testevent.EventBusMock)
	eventConsumerRepo := new(eventstorerepo.EventConsumerRepositoryMock)
	handler := NewHandler(eventBus, eventConsumerRepo)

	eventConsumerId := uuid.New()
	eventConsumerRepo.On("FindById", mock.Anything, eventConsumerId).
		Return(nil, domainerrors.NewEntityNotFoundDomainError("EventConsumer", eventConsumerId))

	cmd := Command{EventConsumerId: eventConsumerId}

	_, err := handler.Handle(context.Background(), &cmd)

	assert.NotNil(t, err)
	assert.Equal(t, coreerror.ErrorCode("EVENT_CONSUMER_NOT_FOUND"), err.(coreerror.DomainError).ErrorCode())
	eventConsumerRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}

func TestCoreTransitFailedEventConsumerToPendingHandler_Handle_InvalidStatusError(t *testing.T) {
	eventBus := new(testevent.EventBusMock)
	eventConsumerRepo := new(eventstorerepo.EventConsumerRepositoryMock)
	handler := NewHandler(eventBus, eventConsumerRepo)

	for _, status := range []eventstore.Status{eventstore.Pending, eventstore.Succeeded} {
		eventConsumer := buildventConsumer(status)
		eventConsumerRepo.On("FindById", mock.Anything, eventConsumer.Id).Return(eventConsumer, nil)

		cmd := Command{EventConsumerId: eventConsumer.Id}

		_, err := handler.Handle(context.Background(), &cmd)

		assert.NotNil(t, err)
		assert.Equal(t, coreerror.ErrorCode("EVENT_CONSUMER_UNEXPECTED_STATUS"), err.(coreerror.DomainError).ErrorCode())
		eventConsumerRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
	}
}

func TestCoreTransitFailedEventConsumerToPendingHandler_Handle_DatabaseError(t *testing.T) {
	eventBus := new(testevent.EventBusMock)
	eventConsumerRepo := new(eventstorerepo.EventConsumerRepositoryMock)
	handler := NewHandler(eventBus, eventConsumerRepo)

	customErr := errors.New("database error")
	eventConsumer := buildventConsumer(eventstore.Failed)
	eventConsumerRepo.On("FindById", mock.Anything, eventConsumer.Id).Return(eventConsumer, nil)
	eventConsumerRepo.On("Save", mock.Anything, mock.Anything).Return(customErr)

	cmd := Command{EventConsumerId: eventConsumer.Id}

	_, err := handler.Handle(context.Background(), &cmd)

	assert.NotNil(t, err)
	assert.ErrorIs(t, err, customErr)
	eventConsumerRepo.AssertExpectations(t)
}

func TestCoreTransitFailedEventConsumerToPendingHandler_Handle_EventPublishError(t *testing.T) {
	eventBus := new(testevent.EventBusMock)
	eventConsumerRepo := new(eventstorerepo.EventConsumerRepositoryMock)
	handler := NewHandler(eventBus, eventConsumerRepo)

	customErr := errors.New("publish error")
	eventConsumer := buildventConsumer(eventstore.Failed)
	eventConsumerRepo.On("FindById", mock.Anything, eventConsumer.Id).Return(eventConsumer, nil)
	eventConsumerRepo.On("Save", mock.Anything, mock.Anything).Return(nil)
	eventBus.On("Publish", mock.Anything, mock.Anything).Return(customErr).Once()

	cmd := Command{EventConsumerId: eventConsumer.Id}

	_, err := handler.Handle(context.Background(), &cmd)

	assert.NotNil(t, err)
	assert.ErrorIs(t, err, customErr)
	eventConsumerRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}
