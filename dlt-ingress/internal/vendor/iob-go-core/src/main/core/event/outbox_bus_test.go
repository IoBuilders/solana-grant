package event

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/domain/eventstore"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/port/repository/eventstore"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/observability/metrics"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/utils"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/test/core/event/model"
	"go.opentelemetry.io/otel/metric/noop"
)

var outboxBus *OutboxBus
var mockEventStoreRepo = new(eventstorerepo.EventStoreRepositoryMock)
var listenerRegistry *ListenerRegistry

type BadEvent struct {
	BaseEvent
	Ch chan int `json:"ch"`
}

func TestMain(m *testing.M) {
	Cleanup()
	os.Exit(m.Run())
}

func Cleanup() {
	mockEventStoreRepo = new(eventstorerepo.EventStoreRepositoryMock)
	listenerRegistry = NewListenerRegistry()

	provider := noop.NewMeterProvider()
	noopMeter := provider.Meter("testMeter")
	metrics.SetDefaultMeter(noopMeter)

	outboxBus = NewOutboxBus(
		mockEventStoreRepo,
		listenerRegistry,
		[]string{},
		metrics.NewRegistry(),
	)

}

func TestCoreOutboxBus_Publish_EventStore_Creation_Error(t *testing.T) {
	event := NewBaseEvent()
	registerDummyListenerWithName(event, strings.Repeat("1", 256))

	err := outboxBus.Publish(context.Background(), event)

	assert.NotNil(t, err)
	mockEventStoreRepo.AssertNotCalled(t, "Save")
	t.Cleanup(Cleanup)
}

func TestCoreOutboxBus_Publish_Success_Memory(t *testing.T) {
	event := NewBaseEvent()
	listener := registerDummyListener(event)
	expectedPayload, _ := json.Marshal(event)

	mockEventStoreRepo.On("Save", mock.Anything, mock.MatchedBy(func(eventStore *eventstore.EventStore) bool {
		assert.Equal(t, reflect.TypeOf(event).String(), eventStore.Type)
		assert.JSONEq(t, string(expectedPayload), string(eventStore.Payload))
		assert.Equal(t, eventstore.Memory, eventStore.PublicationType)
		eventConsumers := eventStore.EventConsumers
		assert.Equal(t, 1, len(eventConsumers))
		assert.Equal(t, listener.Name, eventConsumers[0].Type)
		assert.Equal(t, eventstore.Pending, eventConsumers[0].Status)
		return true
	})).Return(nil).Once()

	err := outboxBus.Publish(context.Background(), event)

	assert.Nil(t, err)
	mockEventStoreRepo.AssertExpectations(t)
	t.Cleanup(Cleanup)
}

func TestCoreOutboxBus_Publish_Success_HTTP(t *testing.T) {
	targetUrl := "api/v1/internal/test"
	event := model.NewContractEvent([]string{targetUrl})
	expectedPayload, _ := json.Marshal(event)

	mockEventStoreRepo.On("Save", mock.Anything, mock.MatchedBy(func(eventStore *eventstore.EventStore) bool {
		assert.Equal(t, reflect.TypeOf(event).String(), eventStore.Type)
		assert.JSONEq(t, string(expectedPayload), string(eventStore.Payload))
		assert.Equal(t, eventstore.HTTP, eventStore.PublicationType)
		eventConsumers := eventStore.EventConsumers
		assert.Equal(t, 1, len(eventConsumers))
		assert.Equal(t, targetUrl, eventConsumers[0].Type)
		assert.Equal(t, eventstore.Pending, eventConsumers[0].Status)
		return true
	})).Return(nil).Once()

	err := outboxBus.Publish(context.Background(), event)

	assert.Nil(t, err)
	mockEventStoreRepo.AssertExpectations(t)
	t.Cleanup(Cleanup)
}

func TestCoreOutboxBus_Publish_NilEvent_Error(t *testing.T) {
	err := outboxBus.Publish(context.Background(), nil)
	assert.Error(t, err)

	mockEventStoreRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
	t.Cleanup(Cleanup)
}

func TestCoreOutboxBus_Publish_MappingError_DoesNotSave(t *testing.T) {
	ev := &BadEvent{
		BaseEvent: *NewBaseEvent(),
		Ch:        make(chan int),
	}

	registerDummyListener(ev)

	err := outboxBus.Publish(context.Background(), ev)
	assert.Error(t, err)

	mockEventStoreRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
	t.Cleanup(Cleanup)
}

func TestCoreOutboxBus_Publish_SaveError_ReturnsError(t *testing.T) {
	ev := NewBaseEvent()
	registerDummyListener(ev)

	mockEventStoreRepo.On("Save", mock.Anything, mock.Anything).
		Return(errors.New("db down")).
		Once()

	err := outboxBus.Publish(context.Background(), ev)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db down")

	mockEventStoreRepo.AssertExpectations(t)
	t.Cleanup(Cleanup)
}

func registerDummyListener(event Event) *ListenerDefinition {
	return registerDummyListenerWithName(event, "dummyListener")
}

func registerDummyListenerWithName(event Event, name string) *ListenerDefinition {
	eventType := utils.GetType(event).String()
	listenerDefinition := NewListenerDefinition(name, func(ctx context.Context, e Event) error {
		print("listening")
		return nil
	})

	listenerRegistry.Register(eventType, listenerDefinition)

	return &listenerDefinition
}
