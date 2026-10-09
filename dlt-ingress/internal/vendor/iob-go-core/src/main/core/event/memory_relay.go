package event

import (
	"context"
	"encoding/json"
	"fmt"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/observability/metrics"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/retry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/panicinfo"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/domain/eventstore"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/port/repository/eventstore"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// DefaultSemaphoreWaitTimeout is the maximum time a goroutine will park waiting
// for a concurrency slot. It must stay well below the janitor threshold so that
// the goroutine always resets the consumer to Pending itself before the janitor
// can see the PROCESSING row and reset it in parallel, which would cause two
// goroutines to process the same consumer concurrently.
const DefaultSemaphoreWaitTimeout = 30 * time.Second

type MemoryRelay struct {
	registry                *ListenerRegistry
	eventConsumerRepository eventstorerepo.EventConsumerRepository
	tracer                  trace.Tracer
	retryer                 retry.Retryer
	retryOptions            retry.Options
	metricsRegistry         *metrics.Registry
	// wg tracks in-flight listener goroutines so the application can drain them
	// on shutdown before the process exits. See Wait.
	wg                   sync.WaitGroup
	semaphoreWaitTimeout time.Duration
}

func NewMemoryRelay(
	registry *ListenerRegistry,
	eventConsumerRepository eventstorerepo.EventConsumerRepository,
	tracer trace.Tracer,
	retryer retry.Retryer,
	retryOptions retry.Options,
	metricsRegistry *metrics.Registry,
) *MemoryRelay {
	return &MemoryRelay{
		registry:                registry,
		eventConsumerRepository: eventConsumerRepository,
		tracer:                  tracer,
		retryer:                 retryer,
		retryOptions:            retryOptions,
		metricsRegistry:         metricsRegistry,
		semaphoreWaitTimeout:    DefaultSemaphoreWaitTimeout,
	}
}

// WithSemaphoreWaitTimeout overrides the maximum time a goroutine may park
// waiting for a concurrency slot. Set this to a value less than the janitor
// threshold to prevent double-delivery of the same consumer.
func (m *MemoryRelay) WithSemaphoreWaitTimeout(d time.Duration) *MemoryRelay {
	m.semaphoreWaitTimeout = d
	return m
}

// Wait blocks until all in-flight listener goroutines have finished. Call it
// from the application shutdown hook after the StoreWorker's context has been
// cancelled, so listeners release their DB connections before the process exits
func (m *MemoryRelay) Wait() {
	m.wg.Wait()
}

func (m *MemoryRelay) Relay(ctx context.Context, events *[]eventstore.EventStore) error {
	for i := range *events {
		if err := ctx.Err(); err != nil {
			return err
		}
		eventStore := &(*events)[i]
		eventInstance, err := m.registry.CreateInstance(eventStore.Type)
		if err != nil {
			// TODO, for now not return an error because cross events use the same event consumer & store entities
			logger.InfoWithCtx(ctx, fmt.Sprintf("Failed to create instance for event type %s: %s", eventStore.Type, err), "error", err)
			continue
		}

		if err := json.Unmarshal(eventStore.Payload, eventInstance); err != nil {
			return fmt.Errorf("failed to unmarshal payload for event type %s: %w", eventStore.Type, err)
		}

		if ev, ok := eventInstance.(Event); ok {
			var listenerDefinitions []ListenerDefinition
			if eventStore.PublicationType == eventstore.Memory {
				listenerDefinitions = m.registry.GetListeners(ev)
			} else {
				listenerDefinitions = make([]ListenerDefinition, len(eventStore.EventConsumers))
				for i, eventConsumer := range eventStore.EventConsumers {
					listenerDefinitions[i] = m.newHTTPListenerDefinition(eventStore, &eventConsumer)
				}
			}
			if err := m.executeListeners(ctx, listenerDefinitions, ev, eventStore); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("event type %s does not implement Event interface", eventStore.Type)
		}
	}
	return nil
}

func (m *MemoryRelay) executeListeners(ctx context.Context, listeners []ListenerDefinition, ev Event, eventStore *eventstore.EventStore) error {
	consumerMap := make(map[string]eventstore.EventConsumer)
	for _, c := range eventStore.EventConsumers {
		consumerMap[c.Type] = c
	}

	carrier := propagation.MapCarrier{"traceparent": eventStore.TraceParent}

	for _, listener := range listeners {
		if err := ctx.Err(); err != nil {
			return err
		}
		consumer, ok := consumerMap[listener.Name]
		if !ok {
			continue
		}

		m.wg.Go(func() {
			// Acquire a concurrency slot before opening a span or touching the DB.
			// Parked goroutines are ~2 KB and do not hold a DB connection, so the
			// dispatch loop is never blocked: other listeners for the same event
			// continue to spawn and execute freely.
			//
			// The wait is bounded by semaphoreWaitTimeout (always < janitor threshold)
			// so the janitor never sees a PROCESSING consumer that is merely parked
			// here and resets it to Pending in parallel, which would cause two
			// goroutines to process the same consumer concurrently.
			waitCtx, waitCancel := context.WithTimeout(ctx, m.semaphoreWaitTimeout)
			defer waitCancel()

			select {
			case listener.semaphore <- struct{}{}:
				defer func() { <-listener.semaphore }()
			case <-waitCtx.Done():
				// Timeout or cancel: reset to Pending so the relay re-dispatches on the next
				// cycle without needing the janitor and without risk of double-delivery.
				m.resetConsumerToPending(consumer, ev)
				return
			}

			var (
				listenerCtx context.Context
				cancel      context.CancelFunc
			)
			if listener.executionTimeout > 0 {
				listenerCtx, cancel = context.WithTimeout(ctx, listener.executionTimeout)
			} else {
				listenerCtx, cancel = context.WithCancel(ctx)
			}
			spanCtx, span := m.tracer.Start(
				otel.GetTextMapPropagator().Extract(listenerCtx, carrier),
				listener.Name,
				trace.WithSpanKind(trace.SpanKindInternal),
			)
			defer span.End()

			defer func() {
				if r := recover(); r != nil {
					m.updateConsumerStatus(spanCtx, consumer, eventstore.Failed, ev, fmt.Errorf("panic: %v", r), r)
				}
				cancel()
			}()
			err := listener.Execute(spanCtx, ev)
			if err != nil {
				m.updateConsumerStatus(spanCtx, consumer, eventstore.Failed, ev, err, nil)
			} else {
				m.updateConsumerStatus(spanCtx, consumer, eventstore.Succeeded, ev, err, nil)
			}
		})
	}
	return nil
}

func (m *MemoryRelay) updateConsumerStatus(
	ctx context.Context,
	c eventstore.EventConsumer,
	status eventstore.Status,
	ev Event,
	err error,
	//Note: recover can be nil since this function is called out of recovery
	recover any) {
	c.Status = status
	if err != nil {
		c.ErrorDetails = new(err.Error())
		marshal, _ := json.Marshal(ev)
		if recover != nil {
			info := panicinfo.NewPanicInfo(recover)
			logger.ErrorWithCtx(
				ctx,
				fmt.Sprintf("%s Error in listener %s for event %s", panicinfo.PANIC_RECOVERED_TAG, c.Type, marshal),
				"error", err,
				"file", info.File,
				"line", info.Line,
			)
		} else {
			logger.ErrorWithCtx(ctx, fmt.Sprintf("Error in listener %s for event %s", c.Type, marshal), "error", err)
		}
		m.registerMetric(ctx)
	}

	var saveCtx context.Context
	if ctx.Err() != nil {
		saveCtx = context.Background() // If we use a context with error, database operations will fail leaving consumers in processing statuses
	} else {
		saveCtx = ctx
	}
	if saveErr := m.eventConsumerRepository.Save(saveCtx, &c); saveErr != nil {
		logger.ErrorWithCtx(ctx, fmt.Sprintf("Critical: Could not save status for %s", c.Type), "error", saveErr)
	}
}

func (m *MemoryRelay) resetConsumerToPending(c eventstore.EventConsumer, ev Event) {
	c.Status = eventstore.Pending
	if saveErr := m.eventConsumerRepository.Save(context.Background(), &c); saveErr != nil {
		marshal, _ := json.Marshal(ev)
		logger.ErrorWithCtx(context.Background(),
			fmt.Sprintf("Critical: Could not reset consumer to pending for %s, event %s", c.Type, marshal),
			"error", saveErr,
		)
	}
}

func (m *MemoryRelay) newHTTPListenerDefinition(eventStore *eventstore.EventStore, eventConsumer *eventstore.EventConsumer) ListenerDefinition {
	return NewListenerDefinition(eventConsumer.Type, func(ctx context.Context, event Event) error {
		_, err := m.retryer.Execute(ctx, m.retryOptions, func(ctx context.Context) (any, error) {
			res, err := http.Post(eventConsumer.Type, "application/json", strings.NewReader(eventStore.Payload.String()))
			if res != nil {
				defer func() {
					_, _ = io.Copy(io.Discard, res.Body)
					_ = res.Body.Close()
				}()
				if res.StatusCode < 200 || res.StatusCode >= 300 {
					errBody, _ := io.ReadAll(io.LimitReader(res.Body, 512))
					httpErr := fmt.Errorf("error calling HTTP API %s for blockchain event. Status Code %d. Response %s", eventConsumer.Type, res.StatusCode, string(errBody))
					if res.StatusCode >= 500 {
						// Server-side failures may be transient; keep retrying.
						return nil, httpErr
					}
					// Any other non-2xx response (1xx/3xx/4xx) is a deterministic rejection; retrying would just repeat it.
					return nil, retry.NewNonRetryableError(httpErr)
				}
			}
			return nil, err
		})
		return err
	})
}

func (m *MemoryRelay) registerMetric(ctx context.Context) {

	counter, registryErr := m.metricsRegistry.GetCounter(metrics.FailuresMetricCounter)
	if registryErr != nil {
		logger.WarnWithCtx(ctx, "failed to get metric", "metricName", metrics.FailuresMetricCounter, "error", registryErr)
	} else {
		attrs := metric.WithAttributes(
			attribute.String(metrics.FailureTypeAttr, "event"),
		)
		counter.Add(ctx, 1, attrs)
	}
}
