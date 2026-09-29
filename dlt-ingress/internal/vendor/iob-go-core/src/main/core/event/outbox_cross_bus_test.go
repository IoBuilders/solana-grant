package event

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/domain/eventstore"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/port/repository/eventstore"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/observability/metrics"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/utils"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/test/core/event/model"
)

type DummyCrossEvent struct {
	BaseEvent
	Message string `json:"message"`
}

type BadCrossEvent struct {
	BaseEvent
	Ch chan int `json:"ch"`
}

func TestCoreMapAnyToEventStore_Memory_UsesRegistryListeners(t *testing.T) {
	reg := NewListenerRegistry()

	ev := &DummyCrossEvent{
		BaseEvent: *NewBaseEvent(),
		Message:   "test",
	}

	listenerKey := utils.GetType(ev).String()

	reg.Register(listenerKey,
		NewListenerDefinition("l1", func(ctx context.Context, e Event) error { return nil }),
		NewListenerDefinition("l2", func(ctx context.Context, e Event) error { return nil }),
	)

	es, err := mapAnyToEventStore(context.Background(), reg, ev, true)
	assert.NoError(t, err)

	expectedPayload, _ := json.Marshal(ev)

	assert.Equal(t, reflect.TypeOf(ev).String(), es.Type)
	assert.JSONEq(t, string(expectedPayload), string(es.Payload))
	assert.Equal(t, eventstore.Memory, es.PublicationType)
	assert.Len(t, es.EventConsumers, 2)
	assert.Equal(t, "l1", es.EventConsumers[0].Type)
	assert.Equal(t, eventstore.Pending, es.EventConsumers[0].Status)
	assert.Equal(t, "l2", es.EventConsumers[1].Type)
	assert.Equal(t, eventstore.Pending, es.EventConsumers[1].Status)
}

func TestCoreMapAnyToEventStore_HTTP_UsesTargetUrlsWhenTypeIsHttpPublication(t *testing.T) {
	reg := NewListenerRegistry()

	targetUrl := "http://example.internal/target"
	ev := model.NewContractEvent([]string{targetUrl})

	eventType := reflect.TypeOf(ev).String()
	assert.Contains(t, HttpPublicationEventTypes, eventType, "test event type must be configured as HTTP publication type")

	es, err := mapAnyToEventStore(context.Background(), reg, ev, true)
	assert.NoError(t, err)

	expectedPayload, _ := json.Marshal(ev)

	assert.Equal(t, eventType, es.Type)
	assert.JSONEq(t, string(expectedPayload), string(es.Payload))
	assert.Equal(t, eventstore.HTTP, es.PublicationType)
	assert.Len(t, es.EventConsumers, 1)
	assert.Equal(t, targetUrl, es.EventConsumers[0].Type)
	assert.Equal(t, eventstore.Pending, es.EventConsumers[0].Status)
}

func TestCoreOutboxCrossBus_Publish_SavesEventStore(t *testing.T) {
	repo := new(eventstorerepo.EventStoreRepositoryMock)
	reg := NewListenerRegistry()
	bus := NewOutboxCrossBus(repo, reg, metrics.NewRegistry())

	ev := &DummyCrossEvent{
		BaseEvent: *NewBaseEvent(),
		Message:   "test",
	}

	listenerKey := utils.GetType(ev).String()
	reg.Register(listenerKey, NewListenerDefinition("dummyListener", func(ctx context.Context, e Event) error { return nil }))

	expectedPayload, _ := json.Marshal(ev)
	expectedType := reflect.TypeOf(ev).String() // stored in EventStore.Type

	repo.On("Save", mock.Anything, mock.MatchedBy(func(es *eventstore.EventStore) bool {
		assert.Equal(t, expectedType, es.Type)
		assert.JSONEq(t, string(expectedPayload), string(es.Payload))
		assert.Equal(t, eventstore.Memory, es.PublicationType)

		if !assert.Len(t, es.EventConsumers, 1) {
			return false
		}
		assert.Equal(t, "dummyListener", es.EventConsumers[0].Type)
		assert.Equal(t, eventstore.Pending, es.EventConsumers[0].Status)
		return true
	})).Return(nil).Once()

	err := bus.Publish(context.Background(), ev)
	assert.NoError(t, err)

	repo.AssertExpectations(t)
}

func TestCoreOutboxCrossBus_Publish_WhenMappingFails_ReturnsErrorAndDoesNotSave(t *testing.T) {
	repo := new(eventstorerepo.EventStoreRepositoryMock)
	reg := NewListenerRegistry()
	bus := NewOutboxCrossBus(repo, reg, metrics.NewRegistry())

	err := bus.Publish(context.Background(), nil)
	assert.Error(t, err)

	repo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}

func TestCoreOutboxCrossBus_Publish_WhenSaveFails_ReturnsError(t *testing.T) {
	repo := new(eventstorerepo.EventStoreRepositoryMock)
	reg := NewListenerRegistry()
	bus := NewOutboxCrossBus(repo, reg, metrics.NewRegistry())

	ev := &DummyCrossEvent{
		BaseEvent: *NewBaseEvent(),
		Message:   "test",
	}

	listenerKey := utils.GetType(ev).String()
	reg.Register(listenerKey, NewListenerDefinition("dummyListener", func(ctx context.Context, e Event) error { return nil }))

	expectedPayload, _ := json.Marshal(ev)
	expectedType := reflect.TypeOf(ev).String()

	repo.On("Save", mock.Anything, mock.MatchedBy(func(es *eventstore.EventStore) bool {
		assert.Equal(t, expectedType, es.Type)
		assert.JSONEq(t, string(expectedPayload), string(es.Payload))
		assert.Equal(t, eventstore.Memory, es.PublicationType)
		return true
	})).Return(errors.New("db down")).Once()

	err := bus.Publish(context.Background(), ev)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db down")

	repo.AssertExpectations(t)
}

func TestCoreOutboxCrossBus_Publish_WhenMapperReturnsError_TriggersMappingErrorBranch(t *testing.T) {
	repo := new(eventstorerepo.EventStoreRepositoryMock)
	reg := NewListenerRegistry()
	bus := NewOutboxCrossBus(repo, reg, metrics.NewRegistry())

	ev := &BadCrossEvent{
		BaseEvent: *NewBaseEvent(),
		Ch:        make(chan int),
	}

	err := bus.Publish(context.Background(), ev)
	assert.Error(t, err)

	repo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}
