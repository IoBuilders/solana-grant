package signandsend

import (
	"context"
	"dlt-ingress/src/main/dltingress/config/metrics"

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

	counter, registryErr := h.registry.GetCounter(dltingressmetrics.DltTransactionsMetric)
	if registryErr != nil {
		logger.WarnWithCtx(ctx, "failed to get metric", "metricName", dltingressmetrics.DltTransactionsMetric, "error", registryErr)
	} else {
		attrs := metric.WithAttributes(
			attribute.Bool("success", err == nil),
			attribute.String("networkId", cmd.NetworkId),
			attribute.String("smartContractId", cmd.SmartContractId),
			attribute.String("smartContractName", cmd.SmartContractName),
			attribute.String("methodName", cmd.MethodName),
		)
		counter.Add(ctx, 1, attrs)
	}

	return res, err
}
