package dltingressmetrics

import (
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/observability/metrics"
)

const (
	DltTransactionsMetric   metrics.MetricName = "dlt.transactions"
	QueueTransactionsMetric metrics.MetricName = "queue.transactions"
)

func SetupMetrics(registry *metrics.Registry) error {
	dltTransactionsCounter, err := metrics.Counter(DltTransactionsMetric, "Number of transactions that are sent to the DLT", "{transaction}")
	if err != nil {
		return err
	}
	if err = registry.RegisterCounter(DltTransactionsMetric, dltTransactionsCounter); err != nil {
		return err
	}

	queueUpDownCounter, err := metrics.UpDownCounter(QueueTransactionsMetric, "Number of transactions that are in queue", "{transaction}")
	if err != nil {
		return err
	}
	if err = registry.RegisterUpDownCounter(QueueTransactionsMetric, queueUpDownCounter); err != nil {
		return err
	}

	return nil
}
