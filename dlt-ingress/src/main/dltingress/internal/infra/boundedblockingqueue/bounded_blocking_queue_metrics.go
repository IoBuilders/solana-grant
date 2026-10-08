package boundedblockingqueue

import (
	"context"
	"dlt-ingress/src/main/dltingress/config/metrics"

	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/observability/metrics"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type BoundedBlockingQueueDbAdapterMetrics struct {
	*BoundedBlockingQueueDbAdapter
	registry *metrics.Registry
}

func NewBoundedBlockingQueueDbAdapterMetrics(BoundedBlockingQueueDbAdapter *BoundedBlockingQueueDbAdapter, registry *metrics.Registry) Port {
	return &BoundedBlockingQueueDbAdapterMetrics{
		BoundedBlockingQueueDbAdapter: BoundedBlockingQueueDbAdapter,
		registry:                      registry,
	}
}

func (a *BoundedBlockingQueueDbAdapterMetrics) Put(ctx context.Context, networkId string, value uuid.UUID) error {
	err := a.BoundedBlockingQueueDbAdapter.Put(ctx, networkId, value)

	counter, registryErr := a.registry.GetUpDownCounter(dltingressmetrics.QueueTransactionsMetric)
	if registryErr != nil {
		logger.WarnWithCtx(ctx, "failed to get metric", "metricName", dltingressmetrics.QueueTransactionsMetric, "error", registryErr)
	} else {
		counter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("networkId", networkId),
			attribute.Bool("success", err == nil),
		))
	}

	return err
}

func (a *BoundedBlockingQueueDbAdapterMetrics) Take(ctx context.Context, networkId string) error {
	err := a.BoundedBlockingQueueDbAdapter.Take(ctx, networkId)

	counter, registryErr := a.registry.GetUpDownCounter(dltingressmetrics.QueueTransactionsMetric)
	if registryErr != nil {
		logger.WarnWithCtx(ctx, "failed to get metric", "metricName", dltingressmetrics.QueueTransactionsMetric, "error", registryErr)
	} else {
		counter.Add(ctx, -1, metric.WithAttributes(
			attribute.String("networkId", networkId),
			attribute.Bool("success", err == nil),
		))
	}

	return err
}
