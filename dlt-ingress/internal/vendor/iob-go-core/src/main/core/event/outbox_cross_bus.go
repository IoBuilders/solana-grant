package event

import (
	"context"
	"fmt"
	"reflect"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/port/repository/eventstore"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/observability/metrics"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type OutboxCrossBus struct {
	repository      eventstorerepo.Repository
	registry        *ListenerRegistry
	metricsRegistry *metrics.Registry
}

func NewOutboxCrossBus(repository eventstorerepo.Repository, registry *ListenerRegistry, metricsRegistry *metrics.Registry) *OutboxCrossBus {
	return &OutboxCrossBus{
		repository:      repository,
		registry:        registry,
		metricsRegistry: metricsRegistry,
	}
}

func (b *OutboxCrossBus) Publish(ctx context.Context, event CrossEvent) error {
	if event == nil {
		return fmt.Errorf("cross event is nil")
	}
	eventType := reflect.TypeOf(event).String()

	eventStore, err := mapAnyToEventStore(ctx, b.registry, event, true)

	if err != nil {
		logger.ErrorWithCtx(ctx, fmt.Sprintf("Failed mapping cross event %s to event store: ", eventType), "error", err)
		return err
	}

	if err := b.repository.Save(ctx, eventStore); err != nil {
		logger.ErrorWithCtx(ctx, fmt.Sprintf("Failed saving event store"), "error", err)
		return err
	}
	logger.InfoWithCtx(ctx, "[CROSS-EVENT-BUS] Event published", "event", eventType)

	counter, errCounter := b.metricsRegistry.GetCounter(metrics.EventsMetricCounter)

	if errCounter != nil {
		logger.WarnWithCtx(ctx, "error on getting metric from registry", "error", errCounter, "eventType", eventType)
	} else {
		counter.Add(ctx, 1, metric.WithAttributes(attribute.Bool("cross", true), attribute.String("eventType", eventType)))
	}

	return nil
}
