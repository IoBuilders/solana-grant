package main

import (
	"context"
	"dlt-ingress"
	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/config/router"
	"dlt-ingress/src/main/core"
	"dlt-ingress/src/main/core/api/auth"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/panicinfo"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/observability"
)

// @title DLT Ingress API
// @version 1.0.0
// @description API for DLT Ingress - Build, sign and send transactions on any DLT or Blockchain.
// @host localhost:8080
// @BasePath /api/v1/
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	originalLoc, _ := time.LoadLocation(time.Local.String())
	systemTime := time.Now().In(originalLoc)

	_ = os.Setenv("TZ", "UTC")

	loc, _ := time.LoadLocation("UTC")
	time.Local = loc

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.InfoWithCtx(ctx, "System local time", "time", systemTime)
	logger.InfoWithCtx(ctx, "Backend local time", "time", time.Now())

	// Load Config
	if err := config.LoadConfig(ctx); err != nil {
		logger.ErrorWithCtx(ctx, "failed to load config", "error", err)
		os.Exit(1)
	}

	// Setup OpenTelemetry SDK
	shutdownOTel, err := observability.SetupOTelSDK(
		ctx,
		observability.WithLogLevel(config.AppConfig.LoggingConfig.Level),
		observability.WithLogFormat(config.AppConfig.LoggingConfig.Format),
		observability.WithScope(observability.NewScope("gitlab.com/iobuilders/projects/eng/o2d/dlt-ingress", dlt_ingress.Version)),
	)
	if err != nil {
		logger.ErrorWithCtx(ctx, "failed to setup otel sdk", "error", err)
		os.Exit(1)
	}
	defer func() {
		otelCtx, otelCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer otelCancel()

		if err := shutdownOTel(otelCtx); err != nil {
			logger.ErrorWithCtx(context.Background(), "failed to shutdown otel sdk", "error", err)
		}
	}()

	// Setup Auth
	if err := auth.SetupAuth(ctx); err != nil {
		logger.ErrorWithCtx(ctx, "failed to setup authorization", "error", err)
		os.Exit(1)
	}

	// Start the application
	app, err := core.StartApplication(ctx)
	if err != nil {
		logger.ErrorWithCtx(ctx, "failed to start application", "error", err)
		os.Exit(1)
	}
	engine, err := router.ConfigRouter(app)
	if err != nil {
		logger.ErrorWithCtx(ctx, "failed to config router", "error", err)
		os.Exit(1)
	}

	// Configure HTTP server manually
	baseUrl := config.AppConfig.Server.Hostname + ":" + config.AppConfig.Server.Port
	config.AppConfig.Server.BaseUrl = "http://" + baseUrl
	srv := &http.Server{
		Addr:              baseUrl,
		Handler:           engine,
		ReadHeaderTimeout: config.AppConfig.Server.ReadHeaderTimeout,
		ReadTimeout:       config.AppConfig.Server.ReadTimeout,
		WriteTimeout:      config.AppConfig.Server.WriteTimeout,
		IdleTimeout:       config.AppConfig.Server.IdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				info := panicinfo.NewPanicInfo(r)
				logger.ErrorWithCtx(
					ctx,
					fmt.Sprintf("%s %s", panicinfo.PANIC_RECOVERED_TAG, "panic while running http server"),
					"panic", info.Message,
					"file", info.File,
					"line", info.Line,
					"stack", info.FormatStackTrace(),
				)

				select {
				case errCh <- err:
				default:
				}
			}
		}()

		logger.InfoWithCtx(ctx, "server started", "url", baseUrl)

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			err := fmt.Errorf("http server failed: %w", err)
			select {
			case errCh <- err:
			default:
			}
		}
	}()

	// Listen for shutdown signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	defer signal.Stop(sigs)

	var exitErr error
	select {
	case sig := <-sigs:
		logger.InfoWithCtx(ctx, "shutdown signal received", "signal", sig.String())
	case err := <-errCh:
		exitErr = err
		logger.ErrorWithCtx(ctx, "server failure detected, shutting down application", "error", err)
	}

	// shutdown HTTP server
	httpCtx, httpCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer httpCancel()
	if err := srv.Shutdown(httpCtx); err != nil {
		logger.ErrorWithCtx(ctx, "failed to shutdown http server", "error", err)
		if exitErr == nil {
			exitErr = err
		}
	}

	// Cancel root context (propagate shutdown to scheduler, etc.)
	cancel()

	// shutdown the application components
	appCtx, appCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer appCancel()
	if err := app.Shutdown(appCtx); err != nil {
		logger.ErrorWithCtx(ctx, "failed to shutdown application", "error", err)
		if exitErr == nil {
			exitErr = err
		}
	}

	logger.InfoWithCtx(ctx, "shutdown complete")

	if exitErr != nil {
		os.Exit(1)
	}
}
