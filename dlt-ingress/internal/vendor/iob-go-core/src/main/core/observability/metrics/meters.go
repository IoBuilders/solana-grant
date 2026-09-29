package metrics

import (
	"fmt"

	"go.opentelemetry.io/otel/metric"
)

var defaultMeter metric.Meter

func SetDefaultMeter(meter metric.Meter) {
	defaultMeter = meter
}

func Counter(metricName MetricName, description string, unit string) (metric.Int64Counter, error) {
	counter, err := defaultMeter.Int64Counter(string(metricName), metric.WithDescription(description), metric.WithUnit(unit))
	if err != nil {
		return nil, fmt.Errorf("error creating counter metric: %w", err)
	}
	return counter, nil
}

func Histogram(metricName MetricName, description string, unit string) (metric.Float64Histogram, error) {
	histogram, err := defaultMeter.Float64Histogram(string(metricName), metric.WithDescription(description), metric.WithUnit(unit))
	if err != nil {
		return nil, fmt.Errorf("error creating histogram metric: %w", err)
	}
	return histogram, nil
}

func Gauge(metricName MetricName, description string, unit string) (metric.Float64Gauge, error) {
	gauge, err := defaultMeter.Float64Gauge(string(metricName), metric.WithDescription(description), metric.WithUnit(unit))
	if err != nil {
		return nil, fmt.Errorf("error creating gauge metric: %w", err)
	}
	return gauge, nil
}

func UpDownCounter(metricName MetricName, description string, unit string) (metric.Int64UpDownCounter, error) {
	c, err := defaultMeter.Int64UpDownCounter(string(metricName), metric.WithDescription(description), metric.WithUnit(unit))
	if err != nil {
		return nil, fmt.Errorf("error creating up down counter metric: %w", err)
	}
	return c, nil
}
