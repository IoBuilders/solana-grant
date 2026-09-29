package event

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/observability/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/utils"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/domain/eventstore"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/port/repository/eventstore"
)

var HttpPublicationEventTypes = []string{"commontransaction.TxRevertEvent", "model.ContractEvent"}

type OutboxBus struct {
	repository                eventstorerepo.Repository
	registry                  *ListenerRegistry
	excludedEventsFromInfoLog []string
	metricsRegistry           *metrics.Registry
}

func NewOutboxBus(repository eventstorerepo.Repository, registry *ListenerRegistry, excludedEventsFromInfoLog []string, metricsRegistry *metrics.Registry) *OutboxBus {
	return &OutboxBus{
		repository:                repository,
		registry:                  registry,
		excludedEventsFromInfoLog: excludedEventsFromInfoLog,
		metricsRegistry:           metricsRegistry,
	}
}

func (eb *OutboxBus) Publish(ctx context.Context, event Event) error {
	if event == nil {
		return fmt.Errorf("event is nil")
	}
	eventType := reflect.TypeOf(event).String()

	eventStore, err := mapAnyToEventStore(ctx, eb.registry, event, false)

	if err != nil {
		logger.ErrorWithCtx(ctx, fmt.Sprintf("Failed mapping event %s to event store: ", eventType), "error", err)
		return err
	}

	if err := eb.repository.Save(ctx, eventStore); err != nil {
		logger.ErrorWithCtx(ctx, fmt.Sprintf("Failed saving event store"), "error", err)
		return err
	}
	logMessage := "[EVENT-BUS] Event published"
	logArgs := []any{"event", eventType}
	if slices.Contains(eb.excludedEventsFromInfoLog, eventType) {
		logger.DebugWithCtx(ctx, logMessage, logArgs...)
	} else {
		logger.InfoWithCtx(ctx, logMessage, logArgs...)
	}

	counter, errCounter := eb.metricsRegistry.GetCounter(metrics.EventsMetricCounter)

	if errCounter != nil {
		logger.WarnWithCtx(ctx, "error on getting metric from registry", "error", errCounter, "eventType", eventType)
	} else {
		counter.Add(ctx, 1, metric.WithAttributes(attribute.Bool("cross", false), attribute.String("eventType", eventType)))
	}

	return nil
}

func mapAnyToEventStore(
	ctx context.Context,
	registry *ListenerRegistry,
	ev any,
	isCross bool,
) (*eventstore.EventStore, error) {
	if ev == nil {
		return nil, fmt.Errorf("event is nil")
	}

	eventType := reflect.TypeOf(ev).String()

	var eventConsumers []eventstore.EventConsumer
	var publicationType eventstore.PublicationType

	if slices.Contains(HttpPublicationEventTypes, eventType) {
		publicationType = eventstore.HTTP

		targetUrls := utils.GetFieldValue(ev, "TargetUrls").([]string)
		eventConsumers = make([]eventstore.EventConsumer, len(targetUrls))

		for i, targetUrl := range targetUrls {
			eventConsumer, err := eventstore.NewEventConsumer(targetUrl, string(eventstore.Pending), isCross)
			if err != nil {
				return nil, err
			}
			eventConsumers[i] = *eventConsumer
		}
	} else {
		publicationType = eventstore.Memory

		listenerKey := utils.GetType(ev).String()
		listeners := registry.GetListenersByEventType(listenerKey)

		eventConsumers = make([]eventstore.EventConsumer, len(listeners))
		for i, listener := range listeners {
			eventConsumer, err := eventstore.NewEventConsumer(listener.Name, string(eventstore.Pending), isCross)
			if err != nil {
				return nil, err
			}
			eventConsumers[i] = *eventConsumer
		}
	}

	eventPayload, err := json.Marshal(ev)
	if err != nil {
		return nil, fmt.Errorf("Event is not marshable")
	}

	return eventstore.NewEventStore(
		eventType,
		eventPayload,
		eventConsumers,
		getTraceParentFromContext(ctx),
		string(publicationType),
		isCross,
	)
}

func getTraceParentFromContext(ctx context.Context) string {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	return carrier["traceparent"]
}
