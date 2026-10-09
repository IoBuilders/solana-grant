package event

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/observability/metrics"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/retry"
	"go.opentelemetry.io/otel"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/domain/eventstore"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/port/repository/eventstore"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/utils"
	"gorm.io/datatypes"
)

// DummyEvent for testing
type DummyEvent struct {
	BaseEvent
	Message string `json:"message"`
}

var dummyTracer = otel.Tracer("dummy-tracer")
var retryer = retry.NewCustomRetryer()
var retryOptions = retry.NewOptions(retry.WithMaxAttempts(2), retry.WithMultiplier(1.0))

func TestCoreMemoryRelay_Relay(t *testing.T) {
	registerHelper := func(registry *ListenerRegistry, eventType string, listeners ...ListenerDefinition) {
		registry.Register(eventType, listeners...)
	}

	dummyEventType := utils.GetType(&DummyEvent{}).String()
	metricsRegistry := metrics.NewRegistry()

	t.Run("Successfully execute listeners and update consumer status to Succeeded for Memory publication type", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			registry := NewListenerRegistry()
			registry.RegisterType("DummyEvent", &DummyEvent{})

			repo := new(eventstorerepo.EventConsumerRepositoryMock)
			relay := NewMemoryRelay(registry, repo, dummyTracer, retryer, retryOptions, metricsRegistry)

			event := DummyEvent{
				BaseEvent: *NewBaseEvent(),
				Message:   "hello",
			}

			listenerExecuted := false
			listener := NewListenerDefinition("test-listener", func(ctx context.Context, ev Event) error {
				listenerExecuted = true
				assert.Equal(t, event.Message, ev.(*DummyEvent).Message)
				return nil
			})

			registerHelper(registry, dummyEventType, listener)

			payload, _ := json.Marshal(event)
			eventStore := eventstore.EventStore{
				Type:            "DummyEvent",
				Payload:         datatypes.JSON(payload),
				PublicationType: eventstore.Memory,
				EventConsumers: []eventstore.EventConsumer{
					{
						Type:   "test-listener",
						Status: eventstore.Pending,
					},
				},
			}

			repo.On("Save", mock.Anything, mock.Anything).Return(nil)

			err := relay.Relay(context.Background(), &[]eventstore.EventStore{eventStore})

			synctest.Wait()
			assert.NoError(t, err)
			assert.True(t, listenerExecuted)

			repo.AssertCalled(t, "Save", mock.Anything, mock.MatchedBy(func(ec *eventstore.EventConsumer) bool {
				return ec.Status == eventstore.Succeeded
			}))
		})
	})

	t.Run("Successfully execute listeners and update consumer status to Succeeded for HTTP publication type", func(t *testing.T) {
		listenerExecuted := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("POST expected, %s received", r.Method)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			listenerExecuted = true
		}))
		defer server.Close()
		registry := NewListenerRegistry()
		registry.RegisterType("DummyEvent", &DummyEvent{})

		repo := new(eventstorerepo.EventConsumerRepositoryMock)
		relay := NewMemoryRelay(registry, repo, dummyTracer, retryer, retryOptions, metricsRegistry)

		event := DummyEvent{
			BaseEvent: *NewBaseEvent(),
			Message:   "hello",
		}

		registerHelper(registry, dummyEventType)

		payload, _ := json.Marshal(event)
		eventStore := eventstore.EventStore{
			Type:            "DummyEvent",
			Payload:         datatypes.JSON(payload),
			PublicationType: eventstore.HTTP,
			EventConsumers: []eventstore.EventConsumer{
				{
					Type:   server.URL,
					Status: eventstore.Pending,
				},
			},
		}

		repo.On("Save", mock.Anything, mock.Anything).Return(nil)

		err := relay.Relay(context.Background(), &[]eventstore.EventStore{eventStore})

		time.Sleep(200 * time.Millisecond)
		assert.NoError(t, err)
		assert.True(t, listenerExecuted)

		repo.AssertCalled(t, "Save", mock.Anything, mock.MatchedBy(func(ec *eventstore.EventConsumer) bool {
			return ec.Status == eventstore.Succeeded
		}))
	})

	t.Run("Fail listener execution and update consumer status to Failed for Memory publication type", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			registry := NewListenerRegistry()
			registry.RegisterType("DummyEvent", &DummyEvent{})

			repo := new(eventstorerepo.EventConsumerRepositoryMock)
			relay := NewMemoryRelay(registry, repo, dummyTracer, retryer, retryOptions, metricsRegistry)

			event := DummyEvent{
				BaseEvent: *NewBaseEvent(),
				Message:   "fail",
			}

			listener := NewListenerDefinition("fail-listener", func(ctx context.Context, ev Event) error {
				return assert.AnError
			})

			registerHelper(registry, dummyEventType, listener)

			payload, _ := json.Marshal(event)
			eventStore := eventstore.EventStore{
				Type:            "DummyEvent",
				Payload:         datatypes.JSON(payload),
				PublicationType: eventstore.Memory,
				EventConsumers: []eventstore.EventConsumer{
					{
						Type:   "fail-listener",
						Status: eventstore.Pending,
					},
				},
			}

			repo.On("Save", mock.Anything, mock.Anything).Return(nil)

			err := relay.Relay(context.Background(), &[]eventstore.EventStore{eventStore})

			synctest.Wait()
			assert.NoError(t, err)
			repo.AssertCalled(t, "Save", mock.Anything, mock.MatchedBy(func(ec *eventstore.EventConsumer) bool {
				return ec.Status == eventstore.Failed && *ec.ErrorDetails == assert.AnError.Error()
			}))
		})
	})

	t.Run("Fail listener execution and update consumer status to Failed for HTTP publication type", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("POST expected, %s received", r.Method)
			}
			attempts++

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()
		registry := NewListenerRegistry()
		registry.RegisterType("DummyEvent", &DummyEvent{})

		repo := new(eventstorerepo.EventConsumerRepositoryMock)
		relay := NewMemoryRelay(registry, repo, dummyTracer, retryer, retryOptions, metricsRegistry)

		event := DummyEvent{
			BaseEvent: *NewBaseEvent(),
			Message:   "fail",
		}

		listener := NewListenerDefinition("fail-listener", func(ctx context.Context, ev Event) error {
			return fmt.Errorf("execution failed")
		})

		registerHelper(registry, dummyEventType, listener)

		payload, _ := json.Marshal(event)
		eventStore := eventstore.EventStore{
			Type:            "DummyEvent",
			Payload:         datatypes.JSON(payload),
			PublicationType: eventstore.HTTP,
			EventConsumers: []eventstore.EventConsumer{
				{
					Type:   server.URL,
					Status: eventstore.Pending,
				},
			},
		}

		repo.On("Save", mock.Anything, mock.Anything).Return(nil)

		err := relay.Relay(context.Background(), &[]eventstore.EventStore{eventStore})

		time.Sleep(2 * time.Second)
		assert.NoError(t, err)
		assert.Equal(t, attempts, 2)
		repo.AssertCalled(t, "Save", mock.Anything, mock.MatchedBy(func(ec *eventstore.EventConsumer) bool {
			return ec.Status == eventstore.Failed && strings.Contains(*ec.ErrorDetails, "error calling HTTP API")
		}))
	})

	t.Run("Fail listener execution without retrying and update consumer status to Failed for a 4xx HTTP response", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()
		registry := NewListenerRegistry()
		registry.RegisterType("DummyEvent", &DummyEvent{})

		repo := new(eventstorerepo.EventConsumerRepositoryMock)
		relay := NewMemoryRelay(registry, repo, dummyTracer, retryer, retryOptions, metricsRegistry)

		event := DummyEvent{
			BaseEvent: *NewBaseEvent(),
			Message:   "fail",
		}

		registerHelper(registry, dummyEventType)

		payload, _ := json.Marshal(event)
		eventStore := eventstore.EventStore{
			Type:            "DummyEvent",
			Payload:         datatypes.JSON(payload),
			PublicationType: eventstore.HTTP,
			EventConsumers: []eventstore.EventConsumer{
				{
					Type:   server.URL,
					Status: eventstore.Pending,
				},
			},
		}

		repo.On("Save", mock.Anything, mock.Anything).Return(nil)

		err := relay.Relay(context.Background(), &[]eventstore.EventStore{eventStore})

		time.Sleep(200 * time.Millisecond)
		assert.NoError(t, err)
		assert.Equal(t, 1, attempts)
		repo.AssertCalled(t, "Save", mock.Anything, mock.MatchedBy(func(ec *eventstore.EventConsumer) bool {
			return ec.Status == eventstore.Failed && strings.Contains(*ec.ErrorDetails, "Status Code 404")
		}))
	})

	t.Run("Fail listener execution without retrying for the upper bound of the 4xx range", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			w.WriteHeader(499)
		}))
		defer server.Close()
		registry := NewListenerRegistry()
		registry.RegisterType("DummyEvent", &DummyEvent{})

		repo := new(eventstorerepo.EventConsumerRepositoryMock)
		relay := NewMemoryRelay(registry, repo, dummyTracer, retryer, retryOptions, metricsRegistry)

		event := DummyEvent{
			BaseEvent: *NewBaseEvent(),
			Message:   "fail",
		}

		registerHelper(registry, dummyEventType)

		payload, _ := json.Marshal(event)
		eventStore := eventstore.EventStore{
			Type:            "DummyEvent",
			Payload:         datatypes.JSON(payload),
			PublicationType: eventstore.HTTP,
			EventConsumers: []eventstore.EventConsumer{
				{
					Type:   server.URL,
					Status: eventstore.Pending,
				},
			},
		}

		repo.On("Save", mock.Anything, mock.Anything).Return(nil)

		err := relay.Relay(context.Background(), &[]eventstore.EventStore{eventStore})

		time.Sleep(200 * time.Millisecond)
		assert.NoError(t, err)
		assert.Equal(t, 1, attempts)
		repo.AssertCalled(t, "Save", mock.Anything, mock.MatchedBy(func(ec *eventstore.EventConsumer) bool {
			return ec.Status == eventstore.Failed && strings.Contains(*ec.ErrorDetails, "Status Code 499")
		}))
	})

	t.Run("Fail listener execution without retrying for a non-2xx, non-4xx HTTP response", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			w.WriteHeader(http.StatusMultipleChoices)
		}))
		defer server.Close()
		registry := NewListenerRegistry()
		registry.RegisterType("DummyEvent", &DummyEvent{})

		repo := new(eventstorerepo.EventConsumerRepositoryMock)
		relay := NewMemoryRelay(registry, repo, dummyTracer, retryer, retryOptions, metricsRegistry)

		event := DummyEvent{
			BaseEvent: *NewBaseEvent(),
			Message:   "fail",
		}

		registerHelper(registry, dummyEventType)

		payload, _ := json.Marshal(event)
		eventStore := eventstore.EventStore{
			Type:            "DummyEvent",
			Payload:         datatypes.JSON(payload),
			PublicationType: eventstore.HTTP,
			EventConsumers: []eventstore.EventConsumer{
				{
					Type:   server.URL,
					Status: eventstore.Pending,
				},
			},
		}

		repo.On("Save", mock.Anything, mock.Anything).Return(nil)

		err := relay.Relay(context.Background(), &[]eventstore.EventStore{eventStore})

		time.Sleep(200 * time.Millisecond)
		assert.NoError(t, err)
		assert.Equal(t, 1, attempts)
		repo.AssertCalled(t, "Save", mock.Anything, mock.MatchedBy(func(ec *eventstore.EventConsumer) bool {
			return ec.Status == eventstore.Failed && strings.Contains(*ec.ErrorDetails, "Status Code 300")
		}))
	})

	t.Run("Handle cases where consumer type doesn't match listener name", func(t *testing.T) {
		registry := NewListenerRegistry()
		registry.RegisterType("DummyEvent", &DummyEvent{})

		repo := new(eventstorerepo.EventConsumerRepositoryMock)
		relay := NewMemoryRelay(registry, repo, dummyTracer, retryer, retryOptions, metricsRegistry)

		event := DummyEvent{
			BaseEvent: *NewBaseEvent(),
		}

		listener := NewListenerDefinition("matched-listener", func(ctx context.Context, ev Event) error {
			return nil
		})

		registerHelper(registry, dummyEventType, listener)

		payload, _ := json.Marshal(event)
		eventStore := eventstore.EventStore{
			Type:            "DummyEvent",
			Payload:         datatypes.JSON(payload),
			PublicationType: eventstore.Memory,
			EventConsumers: []eventstore.EventConsumer{
				{
					Type:   "unmatched-listener",
					Status: eventstore.Pending,
				},
			},
		}

		// Save should NOT be called because listener name "matched-listener" doesn't match consumer type "unmatched-listener"
		// repo.On("Save", mock.Anything, mock.Anything).Return(nil)

		err := relay.Relay(context.Background(), &[]eventstore.EventStore{eventStore})

		assert.NoError(t, err)
		repo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
	})

	t.Run("Handle panic in listener and update consumer status to Failed", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			registry := NewListenerRegistry()
			registry.RegisterType("DummyEvent", &DummyEvent{})

			repo := new(eventstorerepo.EventConsumerRepositoryMock)
			relay := NewMemoryRelay(registry, repo, dummyTracer, retryer, retryOptions, metricsRegistry)

			event := DummyEvent{
				BaseEvent: *NewBaseEvent(),
				Message:   "panic",
			}

			listener := NewListenerDefinition("panic-listener", func(ctx context.Context, ev Event) error {
				panic(assert.AnError)
			})

			registerHelper(registry, dummyEventType, listener)

			payload, _ := json.Marshal(event)
			eventStore := eventstore.EventStore{
				Type:            "DummyEvent",
				Payload:         datatypes.JSON(payload),
				PublicationType: eventstore.Memory,
				EventConsumers: []eventstore.EventConsumer{
					{
						Type:   "panic-listener",
						Status: eventstore.Pending,
					},
				},
			}

			repo.On("Save", mock.Anything, mock.Anything).Return(nil)

			err := relay.Relay(context.Background(), &[]eventstore.EventStore{eventStore})

			synctest.Wait()
			assert.NoError(t, err)
			repo.AssertCalled(t, "Save", mock.Anything, mock.MatchedBy(func(ec *eventstore.EventConsumer) bool {
				return ec.Status == eventstore.Failed && *ec.ErrorDetails == fmt.Sprintf("panic: %v", assert.AnError)
			}))
		})
	})

	t.Run("If context is cancelled the Relay does not process events", func(t *testing.T) {
		registry := NewListenerRegistry()
		registry.RegisterType("DummyEvent", &DummyEvent{})

		repo := new(eventstorerepo.EventConsumerRepositoryMock)
		relay := NewMemoryRelay(registry, repo, dummyTracer, retryer, retryOptions, metricsRegistry)

		event := DummyEvent{
			BaseEvent: *NewBaseEvent(),
			Message:   "hello",
		}

		listenerExecuted := false
		listener := NewListenerDefinition("test-listener", func(ctx context.Context, ev Event) error {
			listenerExecuted = true
			assert.Equal(t, event.Message, ev.(*DummyEvent).Message)
			return nil
		})

		registerHelper(registry, dummyEventType, listener)

		payload, _ := json.Marshal(event)
		eventStore := eventstore.EventStore{
			Type:            "DummyEvent",
			Payload:         datatypes.JSON(payload),
			PublicationType: eventstore.Memory,
			EventConsumers: []eventstore.EventConsumer{
				{
					Type:   "test-listener",
					Status: eventstore.Pending,
				},
			},
		}

		repo.On("Save", mock.Anything, mock.Anything).Return(nil)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := relay.Relay(ctx, &[]eventstore.EventStore{eventStore})

		assert.Errorf(t, err, ctx.Err().Error())
		assert.False(t, listenerExecuted)
		repo.AssertNotCalled(t, "Save")
	})
}
