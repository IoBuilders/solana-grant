package metrics

import (
	"fmt"

	"go.opentelemetry.io/otel/metric"
)

type MetricName string

type Registry struct {
	histograms     map[MetricName]metric.Float64Histogram
	counters       map[MetricName]metric.Int64Counter
	gauges         map[MetricName]metric.Float64Gauge
	upDownCounters map[MetricName]metric.Int64UpDownCounter
}

func NewRegistry() *Registry {
	return &Registry{
		histograms:     make(map[MetricName]metric.Float64Histogram),
		counters:       make(map[MetricName]metric.Int64Counter),
		gauges:         make(map[MetricName]metric.Float64Gauge),
		upDownCounters: make(map[MetricName]metric.Int64UpDownCounter),
	}
}

func (r *Registry) RegisterHistogram(key MetricName, histogram metric.Float64Histogram) error {
	if _, exists := r.histograms[key]; exists {
		return fmt.Errorf("histogram %s already registered", key)
	}
	r.histograms[key] = histogram
	return nil
}

func (r *Registry) RegisterCounter(key MetricName, counter metric.Int64Counter) error {
	if _, exists := r.counters[key]; exists {
		return fmt.Errorf("counter %s already registered", key)
	}
	r.counters[key] = counter
	return nil
}

func (r *Registry) RegisterGauge(key MetricName, gauge metric.Float64Gauge) error {
	if _, exists := r.gauges[key]; exists {
		return fmt.Errorf("gauge %s already registered", key)
	}
	r.gauges[key] = gauge
	return nil
}

func (r *Registry) RegisterUpDownCounter(key MetricName, counter metric.Int64UpDownCounter) error {
	if _, exists := r.upDownCounters[key]; exists {
		return fmt.Errorf("up down counter %s already registered", key)
	}
	r.upDownCounters[key] = counter
	return nil
}

func (r *Registry) GetHistogram(key MetricName) (metric.Float64Histogram, error) {
	histogram, found := r.histograms[key]
	if !found {
		return nil, fmt.Errorf("no histogram registered for metric %s", key)
	}
	return histogram, nil
}

func (r *Registry) GetCounter(key MetricName) (metric.Int64Counter, error) {
	counter, found := r.counters[key]
	if !found {
		return nil, fmt.Errorf("no counter registered for metric %s", key)
	}
	return counter, nil
}

func (r *Registry) GetGauge(key MetricName) (metric.Float64Gauge, error) {
	gauge, found := r.gauges[key]
	if !found {
		return nil, fmt.Errorf("no gauge registered for metric %s", key)
	}
	return gauge, nil
}

func (r *Registry) GetUpDownCounter(key MetricName) (metric.Int64UpDownCounter, error) {
	counter, found := r.upDownCounters[key]
	if !found {
		return nil, fmt.Errorf("no up down counter registered for metric %s", key)
	}
	return counter, nil
}
