package savefailedtransaction

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/observability/metrics"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type CommandHandlerMetrics struct {
	*CommandHandler
	registry *metrics.Registry
}

func NewCommandHandlerMetrics(registry *metrics.Registry, commandHandler *CommandHandler) *CommandHandlerMetrics {
	return &CommandHandlerMetrics{
		registry:       registry,
		CommandHandler: commandHandler,
	}
}

func (h *CommandHandlerMetrics) Handle(ctx context.Context, cmd *Command) (*Response, error) {
	res, err := h.CommandHandler.Handle(ctx, cmd)

	counter, registryErr := h.registry.GetCounter(metrics.FailuresMetricCounter)
	if registryErr != nil {
		logger.WarnWithCtx(ctx, "failed to get metric", "metricName", metrics.FailuresMetricCounter, "error", registryErr)
	} else {
		attrs := metric.WithAttributes(
			attribute.String(metrics.FailureTypeAttr, "transaction"),
		)
		counter.Add(ctx, 1, attrs)
	}

	return res, err
}
