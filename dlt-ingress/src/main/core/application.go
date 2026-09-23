package core

import (
	"context"
	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/core/blockchain/client"
	"dlt-ingress/src/main/core/shared"
	"dlt-ingress/src/main/dltingress/config"
	"dlt-ingress/src/main/dltingress/port/evm"
	"errors"
	"sync"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/cache"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/health"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/health/checks"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/observability/metrics"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/retry"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/query"
	"go.opentelemetry.io/otel"
)

// Application holds the application components
type Application struct {
	Core       *shared.CoreCommon
	DltIngress *dltingressconfig.LazyDltIngress
	// shutdown management
	stopFns []func(context.Context) error
	wg      *sync.WaitGroup
}

var App = &Application{}

type AppSettings struct {
	dltIngressOpts []dltingressconfig.DltIngressOption
}

type AppOption func(settings *AppSettings)

func WithDltIngressOptions(options ...dltingressconfig.DltIngressOption) AppOption {
	return func(settings *AppSettings) {
		settings.dltIngressOpts = options
	}
}

func StartApplication(ctx context.Context, opts ...AppOption) (*Application, error) {
	wg := &sync.WaitGroup{}
	settings := &AppSettings{}
	for _, opt := range opts {
		opt(settings)
	}

	// Initialize dependencies
	tracer := otel.GetTracerProvider().Tracer(config.AppConfig.Application.Name)
	queryBus := query.NewQueryBus()
	retryer := retry.NewCustomRetryer()
	blockchainClient := clientblockchain.NewBlockchainService(config.AppConfig.DltIngress.Networks[0].Url) //nolint:forbidigo  // Will be replaced by naryo-go in the future
	evmClientRegistry := evm.NewClientRegistry()
	healthRegistry := health.NewRegistry()
	healthRegistry.Register(healthchecks.NewEthereumNodeChecker("dlt_node", blockchainClient.GetClient()))
	c := cache.NewGoCache(config.AppConfig.Cache.Caches)
	metricsRegistry := metrics.NewRegistry()

	if err := metrics.CoreMetricsSetup(metricsRegistry); err != nil {
		return nil, err
	}

	core := &shared.CoreCommon{
		QueryBus:          queryBus,
		BlockchainClient:  blockchainClient,
		EvmClientRegistry: evmClientRegistry,
		Retryer:           retryer,
		Tracer:            tracer,
		HealthRegistry:    healthRegistry,
		Cache:             c,
		MetricsRegistry:   metricsRegistry,
	}

	dltIngressModule := dltingressconfig.NewLazyDltIngress(core, settings.dltIngressOpts...)
	dltIngressModule.Init(ctx)

	core.CrossCommandBus = dltIngressModule.CrossCommandBus
	core.CrossQueryBus = dltIngressModule.CrossQueryBus
	core.CrossEventBus = dltIngressModule.CrossEventBus
	core.CrossListenerRegistry = dltIngressModule.CrossListenerRegistry

	// Create all lazy modules
	// Routes are registered against these lazy structs before ListenAndServe.
	App = &Application{
		Core:       core,
		wg:         wg,
		DltIngress: dltIngressModule,
	}

	App.AddStopFn(App.DltIngress.ShutdownFn())

	return App, nil
}

func (a *Application) AddStopFn(fn func(context.Context) error) {
	a.stopFns = append(a.stopFns, fn)
}

func (a *Application) Shutdown(ctx context.Context) error {
	var errs []error

	// stop in reverse order
	for i := len(a.stopFns) - 1; i >= 0; i-- {
		if err := a.stopFns[i](ctx); err != nil {
			errs = append(errs, err)
			logger.ErrorWithCtx(ctx, "stop function failed", "function", a.stopFns[i], "error", err)
		}
	}

	return errors.Join(errs...)
}
