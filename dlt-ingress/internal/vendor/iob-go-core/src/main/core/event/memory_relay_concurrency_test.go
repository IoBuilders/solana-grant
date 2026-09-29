package event

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/domain/eventstore"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/port/repository/eventstore"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/observability/metrics"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/utils"
	"gorm.io/datatypes"
)

// memoryEventBatch builds n Memory-publication EventStores of DummyEvent, each
// carrying a single consumer whose Type matches listenerName so the relay
// dispatches one goroutine per event for that listener.
func memoryEventBatch(listenerName string, n int) []eventstore.EventStore {
	events := make([]eventstore.EventStore, n)
	for i := range events {
		payload, _ := json.Marshal(DummyEvent{BaseEvent: *NewBaseEvent(), Message: "msg"})
		events[i] = eventstore.EventStore{
			Type:            "DummyEvent",
			Payload:         datatypes.JSON(payload),
			PublicationType: eventstore.Memory,
			EventConsumers: []eventstore.EventConsumer{
				{Type: listenerName, Status: eventstore.Pending},
			},
		}
	}
	return events
}

func TestNewListenerDefinition_ConcurrencyLimit(t *testing.T) {
	noop := func(context.Context, Event) error { return nil }

	t.Run("positive limit creates a buffered semaphore of that capacity", func(t *testing.T) {
		l := NewListenerDefinition("limited", noop, WithConcurrencyLimit(3))
		assert.NotNil(t, l.semaphore)
		assert.Equal(t, 3, cap(l.semaphore))
	})

	t.Run("zero limit applies the default cap", func(t *testing.T) {
		l := NewListenerDefinition("defaulted", noop, WithConcurrencyLimit(0))
		assert.NotNil(t, l.semaphore)
		assert.Equal(t, DefaultListenerMaxConcurrency, cap(l.semaphore))
	})

	t.Run("negative limit applies the default cap", func(t *testing.T) {
		l := NewListenerDefinition("defaulted-via-negative", noop, WithConcurrencyLimit(-1))
		assert.NotNil(t, l.semaphore)
		assert.Equal(t, DefaultListenerMaxConcurrency, cap(l.semaphore))
	})

	t.Run("no options applies the default cap", func(t *testing.T) {
		l := NewListenerDefinition("plain", noop)
		assert.NotNil(t, l.semaphore)
		assert.Equal(t, DefaultListenerMaxConcurrency, cap(l.semaphore))
	})
}

func TestNewListenerDefinition_WithOptions(t *testing.T) {
	noop := func(context.Context, Event) error { return nil }

	t.Run("no options applies defaults", func(t *testing.T) {
		l := NewListenerDefinition("l", noop)
		assert.Equal(t, DefaultListenerMaxConcurrency, cap(l.semaphore))
		assert.Equal(t, time.Duration(0), l.executionTimeout)
	})

	t.Run("WithConcurrencyLimit sets the semaphore capacity", func(t *testing.T) {
		l := NewListenerDefinition("l", noop, WithConcurrencyLimit(5))
		assert.Equal(t, 5, cap(l.semaphore))
	})

	t.Run("WithConcurrencyLimit zero falls back to default", func(t *testing.T) {
		l := NewListenerDefinition("l", noop, WithConcurrencyLimit(0))
		assert.Equal(t, DefaultListenerMaxConcurrency, cap(l.semaphore))
	})

	t.Run("WithExecutionTimeout option sets the timeout", func(t *testing.T) {
		l := NewListenerDefinition("l", noop, WithExecutionTimeout(30*time.Second))
		assert.Equal(t, 30*time.Second, l.executionTimeout)
	})

	t.Run("both options applied together", func(t *testing.T) {
		l := NewListenerDefinition("l", noop, WithConcurrencyLimit(3), WithExecutionTimeout(10*time.Second))
		assert.Equal(t, 3, cap(l.semaphore))
		assert.Equal(t, 10*time.Second, l.executionTimeout)
	})
}

func TestListenerDefinition_WithExecutionTimeout(t *testing.T) {
	noop := func(context.Context, Event) error { return nil }

	t.Run("option sets the timeout", func(t *testing.T) {
		l := NewListenerDefinition("l", noop, WithExecutionTimeout(5*time.Second))
		assert.Equal(t, 5*time.Second, l.executionTimeout)
	})

	t.Run("no option leaves timeout at zero", func(t *testing.T) {
		l := NewListenerDefinition("l", noop)
		assert.Equal(t, time.Duration(0), l.executionTimeout)
	})
}

func TestCoreMemoryRelay_ExecutionTimeout(t *testing.T) {
	dummyEventType := utils.GetType(&DummyEvent{}).String()
	metricsRegistry := metrics.NewRegistry()

	t.Run("listener context is cancelled with DeadlineExceeded after the execution timeout", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			registry := NewListenerRegistry()
			registry.RegisterType("DummyEvent", &DummyEvent{})
			repo := new(eventstorerepo.EventConsumerRepositoryMock)
			repo.On("Save", mock.Anything, mock.Anything).Return(nil)
			relay := NewMemoryRelay(registry, repo, dummyTracer, retryer, retryOptions, metricsRegistry)

			var capturedErr error
			listener := NewListenerDefinition("timeout-listener", func(ctx context.Context, ev Event) error {
				<-ctx.Done()
				capturedErr = ctx.Err()
				return capturedErr
			}, WithExecutionTimeout(100*time.Millisecond))
			registry.Register(dummyEventType, listener)

			events := memoryEventBatch("timeout-listener", 1)
			err := relay.Relay(context.Background(), &events)
			assert.NoError(t, err)

			// relay.Wait() blocks the test goroutine, which together with the listener
			// goroutine blocking on <-ctx.Done() allows sync test to advance fake time
			// past the 100ms deadline and fire the cancellation.
			relay.Wait()

			assert.Equal(t, context.DeadlineExceeded, capturedErr)
			repo.AssertNumberOfCalls(t, "Save", 1)
		})
	})

	t.Run("zero timeout leaves execution unbounded", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			registry := NewListenerRegistry()
			registry.RegisterType("DummyEvent", &DummyEvent{})
			repo := new(eventstorerepo.EventConsumerRepositoryMock)
			repo.On("Save", mock.Anything, mock.Anything).Return(nil)
			relay := NewMemoryRelay(registry, repo, dummyTracer, retryer, retryOptions, metricsRegistry)

			release := make(chan struct{})
			var executed atomic.Bool
			// No WithExecutionTimeout — execution should only complete when we release.
			listener := NewListenerDefinition("no-timeout-listener", func(ctx context.Context, ev Event) error {
				<-release
				executed.Store(true)
				return nil
			})
			registry.Register(dummyEventType, listener)

			events := memoryEventBatch("no-timeout-listener", 1)
			err := relay.Relay(context.Background(), &events)
			assert.NoError(t, err)

			synctest.Wait()
			assert.False(t, executed.Load(), "listener must not have completed before release")

			close(release)
			relay.Wait()
			assert.True(t, executed.Load())
		})
	})
}

func TestCoreMemoryRelay_PerListenerConcurrency(t *testing.T) {
	dummyEventType := utils.GetType(&DummyEvent{}).String()
	metricsRegistry := metrics.NewRegistry()

	t.Run("never runs more than maxConcurrency executions of a limited listener at once", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			const limit = 2
			const batch = 5

			registry := NewListenerRegistry()
			registry.RegisterType("DummyEvent", &DummyEvent{})
			repo := new(eventstorerepo.EventConsumerRepositoryMock)
			repo.On("Save", mock.Anything, mock.Anything).Return(nil)
			relay := NewMemoryRelay(registry, repo, dummyTracer, retryer, retryOptions, metricsRegistry)

			var active, maxActive, executed atomic.Int32
			release := make(chan struct{})

			listener := NewListenerDefinition("limited-listener", func(ctx context.Context, ev Event) error {
				n := active.Add(1)
				for { // record the high-water mark of concurrent executions
					m := maxActive.Load()
					if n <= m || maxActive.CompareAndSwap(m, n) {
						break
					}
				}
				<-release // hold the slot until the test releases it
				active.Add(-1)
				executed.Add(1)
				return nil
			}, WithConcurrencyLimit(limit))
			registry.Register(dummyEventType, listener)

			events := memoryEventBatch("limited-listener", batch)
			err := relay.Relay(context.Background(), &events)
			assert.NoError(t, err)

			// All goroutines are now blocked: `limit` inside the listener on
			// <-release, the remaining batch-limit parked on the semaphore send.
			synctest.Wait()
			assert.Equal(t, int32(limit), active.Load(), "exactly maxConcurrency should be running")
			assert.Equal(t, int32(limit), maxActive.Load())

			close(release) // let everything drain in waves of at most `limit`
			synctest.Wait()

			assert.Equal(t, int32(batch), executed.Load(), "every event must still be processed")
			assert.Equal(t, int32(limit), maxActive.Load(), "the cap must never be exceeded")
		})
	})

	t.Run("a listener created without an explicit limit is capped at the default", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			batch := DefaultListenerMaxConcurrency + 5

			registry := NewListenerRegistry()
			registry.RegisterType("DummyEvent", &DummyEvent{})
			repo := new(eventstorerepo.EventConsumerRepositoryMock)
			repo.On("Save", mock.Anything, mock.Anything).Return(nil)
			relay := NewMemoryRelay(registry, repo, dummyTracer, retryer, retryOptions, metricsRegistry)

			var active, maxActive, executed atomic.Int32
			release := make(chan struct{})

			// No explicit limit -> DefaultListenerMaxConcurrency applies.
			listener := NewListenerDefinition("default-listener", func(ctx context.Context, ev Event) error {
				n := active.Add(1)
				for {
					m := maxActive.Load()
					if n <= m || maxActive.CompareAndSwap(m, n) {
						break
					}
				}
				<-release
				active.Add(-1)
				executed.Add(1)
				return nil
			})
			registry.Register(dummyEventType, listener)

			events := memoryEventBatch("default-listener", batch)
			err := relay.Relay(context.Background(), &events)
			assert.NoError(t, err)

			synctest.Wait()
			assert.Equal(t, int32(DefaultListenerMaxConcurrency), active.Load())

			close(release)
			synctest.Wait()
			assert.Equal(t, int32(batch), executed.Load())
			assert.Equal(t, int32(DefaultListenerMaxConcurrency), maxActive.Load(), "the default cap must never be exceeded")
		})
	})

	t.Run("a parked goroutine exits cleanly when the context is cancelled and leaves the consumer for re-delivery", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			registry := NewListenerRegistry()
			registry.RegisterType("DummyEvent", &DummyEvent{})
			repo := new(eventstorerepo.EventConsumerRepositoryMock)
			repo.On("Save", mock.Anything, mock.Anything).Return(nil)
			relay := NewMemoryRelay(registry, repo, dummyTracer, retryer, retryOptions, metricsRegistry)

			var started, executed atomic.Int32
			release := make(chan struct{})

			listener := NewListenerDefinition("limited-listener", func(ctx context.Context, ev Event) error {
				started.Add(1)
				<-release
				executed.Add(1)
				return nil
			}, WithConcurrencyLimit(1))
			registry.Register(dummyEventType, listener)

			events := memoryEventBatch("limited-listener", 2)
			ctx, cancel := context.WithCancel(context.Background())
			err := relay.Relay(ctx, &events)
			assert.NoError(t, err)

			// One goroutine holds the single slot; the second is parked on the
			// semaphore. Only one has started executing.
			synctest.Wait()
			assert.Equal(t, int32(1), started.Load())

			// Cancelling drops the parked goroutine before it ever runs.
			cancel()
			synctest.Wait()
			assert.Equal(t, int32(1), started.Load(), "the parked goroutine must not start after cancellation")

			close(release)
			synctest.Wait()
			assert.Equal(t, int32(1), executed.Load(), "only the goroutine that acquired the slot completes")

			// Save is called twice: once by resetConsumerToPending for the parked
			// goroutine (immediately reset to Pending for re-delivery) and once by
			// updateConsumerStatus for the goroutine that ran to completion.
			repo.AssertNumberOfCalls(t, "Save", 2)
		})
	})

	t.Run("Wait blocks until in-flight listener goroutines finish", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			registry := NewListenerRegistry()
			registry.RegisterType("DummyEvent", &DummyEvent{})
			repo := new(eventstorerepo.EventConsumerRepositoryMock)
			repo.On("Save", mock.Anything, mock.Anything).Return(nil)
			relay := NewMemoryRelay(registry, repo, dummyTracer, retryer, retryOptions, metricsRegistry)

			release := make(chan struct{})
			listener := NewListenerDefinition("drain-listener", func(ctx context.Context, ev Event) error {
				<-release
				return nil
			})
			registry.Register(dummyEventType, listener)

			events := memoryEventBatch("drain-listener", 3)
			err := relay.Relay(context.Background(), &events)
			assert.NoError(t, err)

			waitReturned := make(chan struct{})
			go func() {
				relay.Wait()
				close(waitReturned)
			}()

			// Listeners are blocked on <-release, so Wait must still be blocked.
			synctest.Wait()
			select {
			case <-waitReturned:
				t.Fatal("Wait returned before listeners finished")
			default:
			}

			close(release)
			synctest.Wait()
			select {
			case <-waitReturned:
			default:
				t.Fatal("Wait did not return after listeners finished")
			}
		})
	})
}
