package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric/noop"
)

func TestCoreMetricsRegistry_RegisterHistogram_Successful(t *testing.T) {
	r := NewRegistry()
	err := r.RegisterHistogram("latency", noop.Float64Histogram{})
	require.NoError(t, err)
}

func TestCoreMetricsRegistry_RegisterHistogram_Duplicate_Key_Error(t *testing.T) {
	r := NewRegistry()
	_ = r.RegisterHistogram("latency", noop.Float64Histogram{})
	err := r.RegisterHistogram("latency", noop.Float64Histogram{})
	assert.ErrorContains(t, err, "histogram latency already registered")
}

func TestCoreMetricsRegistry_RegisterCounter_Successful(t *testing.T) {
	r := NewRegistry()
	err := r.RegisterCounter("requests", noop.Int64Counter{})
	require.NoError(t, err)
}

func TestCoreMetricsRegistry_RegisterCounter_Duplicate_Key_Error(t *testing.T) {
	r := NewRegistry()
	_ = r.RegisterCounter("requests", noop.Int64Counter{})
	err := r.RegisterCounter("requests", noop.Int64Counter{})
	assert.ErrorContains(t, err, "counter requests already registered")
}

func TestCoreMetricsRegistry_GetHistogram_Successful(t *testing.T) {
	r := NewRegistry()
	h := noop.Float64Histogram{}
	_ = r.RegisterHistogram("latency", h)
	got, err := r.GetHistogram("latency")
	require.NoError(t, err)
	assert.Equal(t, h, got)

}

func TestCoreMetricsRegistry_GetHistogram_Unkown_Key_Error(t *testing.T) {
	r := NewRegistry()
	_, err := r.GetHistogram("unknown")
	assert.ErrorContains(t, err, "no histogram registered for metric unknown")
}

func TestCoreMetricsRegistry_GetCounter_Successful(t *testing.T) {
	r := NewRegistry()
	c := noop.Int64Counter{}
	_ = r.RegisterCounter("requests", c)
	got, err := r.GetCounter("requests")
	require.NoError(t, err)
	assert.Equal(t, c, got)
}

func TestCoreMetricsRegistry_GetCounter_Unkown_Key_Error(t *testing.T) {
	r := NewRegistry()
	_, err := r.GetCounter("unknown")
	assert.ErrorContains(t, err, "no counter registered for metric unknown")
}

func TestCoreMetricsRegistry_RegisterGauge_Successful(t *testing.T) {
	r := NewRegistry()
	err := r.RegisterGauge("queued_dlt_transactions", noop.Float64Gauge{})
	require.NoError(t, err)
}

func TestCoreMetricsRegistry_RegisterGauge_Duplicate_Key_Error(t *testing.T) {
	r := NewRegistry()
	_ = r.RegisterGauge("queued_dlt_transactions", noop.Float64Gauge{})
	err := r.RegisterGauge("queued_dlt_transactions", noop.Float64Gauge{})
	assert.ErrorContains(t, err, "gauge queued_dlt_transactions already registered")
}

func TestCoreMetricsRegistry_GetGauge_Successful(t *testing.T) {
	r := NewRegistry()
	g := noop.Float64Gauge{}
	_ = r.RegisterGauge("queued_dlt_transactions", g)
	got, err := r.GetGauge("queued_dlt_transactions")
	require.NoError(t, err)
	assert.Equal(t, g, got)
}

func TestCoreMetricsRegistry_GetGauge_Unknown_Key_Error(t *testing.T) {
	r := NewRegistry()
	_, err := r.GetGauge("unknown")
	assert.ErrorContains(t, err, "no gauge registered for metric unknown")
}

func TestCoreMetricsRegistry_RegisterUpDownCounter_Successful(t *testing.T) {
	r := NewRegistry()
	err := r.RegisterUpDownCounter("queued_dlt_transactions", noop.Int64UpDownCounter{})
	require.NoError(t, err)
}

func TestCoreMetricsRegistry_RegisterUpDownCounter_Duplicate_Key_Error(t *testing.T) {
	r := NewRegistry()
	_ = r.RegisterUpDownCounter("queued_dlt_transactions", noop.Int64UpDownCounter{})
	err := r.RegisterUpDownCounter("queued_dlt_transactions", noop.Int64UpDownCounter{})
	assert.ErrorContains(t, err, "up down counter queued_dlt_transactions already registered")
}

func TestCoreMetricsRegistry_GetUpDownCounter_Successful(t *testing.T) {
	r := NewRegistry()
	c := noop.Int64UpDownCounter{}
	_ = r.RegisterUpDownCounter("queued_dlt_transactions", c)
	got, err := r.GetUpDownCounter("queued_dlt_transactions")
	require.NoError(t, err)
	assert.Equal(t, c, got)
}

func TestCoreMetricsRegistry_GetUpDownCounter_Unknown_Key_Error(t *testing.T) {
	r := NewRegistry()
	_, err := r.GetUpDownCounter("unknown")
	assert.ErrorContains(t, err, "no up down counter registered for metric unknown")
}
