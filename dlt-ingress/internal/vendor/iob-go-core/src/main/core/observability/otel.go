package observability

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/observability/metrics"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/contrib/instrumentation/host"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"

	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.39.0"
)

type OTELConfig struct {
	logLevel  string
	logFormat string
	scope     Scope
}

type Option func(*OTELConfig)

type Scope struct {
	ScopeName    string
	ScopeVersion string
}

func NewScope(name string, version string) Scope {
	return Scope{
		ScopeName:    name,
		ScopeVersion: version,
	}
}

func WithLogLevel(logLevel string) Option {
	return func(c *OTELConfig) { c.logLevel = logLevel }
}

func WithLogFormat(logFormat string) Option {
	return func(c *OTELConfig) { c.logFormat = logFormat }
}

func WithScope(scope Scope) Option {
	return func(c *OTELConfig) { c.scope = scope }
}

func SetupOTelSDK(ctx context.Context, opts ...Option) (func(context.Context) error, error) {
	otelConfig := OTELConfig{
		logLevel: "INFO",
		scope:    Scope{},
	}

	for _, opt := range opts {
		if opt != nil {
			opt(&otelConfig)
		}
	}

	var shutdownFuncs []func(context.Context) error

	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Resource
	res, err := resource.New(ctx, resource.WithFromEnv())
	if err != nil {
		return nil, fmt.Errorf("otel resource.New: %w", err)
	}

	// Traces
	shutdownFunc, err := configTracerProvider(ctx, res)
	if err != nil {
		return nil, err
	}
	shutdownFuncs = append(shutdownFuncs, shutdownFunc)

	// Logs
	shutdownFunc, err = configLogProvider(ctx, res)
	if err != nil {
		return nil, err
	}
	shutdownFuncs = append(shutdownFuncs, shutdownFunc)

	// Metrics
	shutdownFunc, err = configMeterProvider(ctx, res)
	if err != nil {
		return nil, err
	}
	shutdownFuncs = append(shutdownFuncs, shutdownFunc)

	// Logger
	configLogger(ctx, otelConfig)

	// Meter
	configMeter(otelConfig)

	return shutdown, nil
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func configTracerProvider(ctx context.Context, res *resource.Resource) (func(context.Context) error, error) {
	spanExp, err := autoexport.NewSpanExporter(ctx)
	if err != nil {
		return nil, fmt.Errorf("otel autoexport span exporter: %w", err)
	}
	var spanProcessor trace.SpanProcessor
	if isConsoleTraceExporter(spanExp) {
		spanProcessor = trace.NewSimpleSpanProcessor(spanExp)
	} else {
		spanProcessor = trace.NewBatchSpanProcessor(spanExp)
	}
	tracerProvider := trace.NewTracerProvider(
		trace.WithResource(res),
		trace.WithSpanProcessor(spanProcessor),
	)
	otel.SetTracerProvider(tracerProvider)
	return tracerProvider.Shutdown, nil
}

func configLogProvider(ctx context.Context, res *resource.Resource) (func(context.Context) error, error) {
	logExp, err := autoexport.NewLogExporter(ctx)
	if err != nil {
		return nil, fmt.Errorf("otel autoexport log exporter: %w", err)
	}
	var processor log.Processor
	var additionalConsoleProcessor log.Processor
	if isConsoleLogExporter(logExp) {
		processor = log.NewSimpleProcessor(logExp)
	} else {
		processor = log.NewBatchProcessor(logExp)
		logExp, err := stdoutlog.New()
		if err != nil {
			return nil, fmt.Errorf("otel stdoutlog log exporter: %w", err)
		}
		// Add additional processor with console log exporter to have always 1 console log exporter
		additionalConsoleProcessor = log.NewSimpleProcessor(logExp)
	}
	loggerProviderOptions := []log.LoggerProviderOption{
		log.WithProcessor(processor),
		log.WithResource(res),
	}
	if additionalConsoleProcessor != nil {
		loggerProviderOptions = append(loggerProviderOptions, log.WithProcessor(additionalConsoleProcessor))
	}
	loggerProvider := log.NewLoggerProvider(loggerProviderOptions...)
	global.SetLoggerProvider(loggerProvider)
	return loggerProvider.Shutdown, nil
}

func configMeterProvider(ctx context.Context, res *resource.Resource) (func(context.Context) error, error) {
	metricReader, err := autoexport.NewMetricReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("otel autoexport metric reader: %w", err)
	}
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(metricReader),
	)
	otel.SetMeterProvider(meterProvider)

	err = runtime.Start()
	if err != nil {
		return nil, fmt.Errorf("otel runtime metrics start error: %w", err)
	}

	err = host.Start()
	if err != nil {
		return nil, fmt.Errorf("otel host metrics start error: %w", err)
	}

	return meterProvider.Shutdown, nil
}

func configLogger(ctx context.Context, otelConfig OTELConfig) {
	slogLogLevel, err := parseLogLevel(otelConfig.logLevel)
	var handler slog.Handler
	if otelConfig.logFormat == "text" {
		handler = NewOtelSlogTextHandler(slogLogLevel)
	} else {
		handler = NewOtelLogLevelHandler(
			slogLogLevel,
			otelConfig.scope,
			otelslog.WithVersion(otelConfig.scope.ScopeVersion),
			otelslog.WithSchemaURL(semconv.SchemaURL),
		)
	}
	slogLogger := slog.New(handler)
	slog.SetDefault(slogLogger)
	logger.SetDefaultLogger(logger.NewSlogLogger(slogLogger))
	if err != nil {
		logger.WarnWithCtx(ctx, fmt.Sprintf("Invalid log level %s. Default log level INFO is configured instead", otelConfig.logLevel))
	}
}

func configMeter(otelConfig OTELConfig) {
	metrics.SetDefaultMeter(
		otel.GetMeterProvider().Meter(
			otelConfig.scope.ScopeName,
			metric.WithInstrumentationVersion(otelConfig.scope.ScopeVersion),
			metric.WithSchemaURL(semconv.SchemaURL),
		),
	)
}

func parseLogLevel(levelStr string) (slog.Level, error) {
	var level slog.Level
	err := level.UnmarshalText([]byte(levelStr))
	if err != nil {
		return slog.LevelInfo, fmt.Errorf("otel stdoutlog log exporter: %w", err)
	}
	return level, nil
}

func isConsoleLogExporter(e log.Exporter) bool {
	_, ok := e.(*stdoutlog.Exporter)
	return ok
}

func isConsoleTraceExporter(e trace.SpanExporter) bool {
	_, ok := e.(*stdouttrace.Exporter)
	return ok
}
