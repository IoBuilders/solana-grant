package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"net/http"
	"os"

	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-http/infrastructure/setup"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/broadcast"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/logging"
	coreconfig "gitlab.com/iobuilders/projects/eng/naryo-go/core/infrastructure/config"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/infrastructure/setup"
	"gitlab.com/iobuilders/projects/eng/naryo-go/persistence-gorm/infrastructure/setup"
	"gitlab.com/iobuilders/projects/eng/naryo-go/server/infrastructure/config"
)

// defaultConfigYAML is the "default application.yaml": the base configuration used unless an
// override file is given (see resolveConfigPath below).
//
//go:embed resources/application.yml
var defaultConfigYAML string

const rootProperty = "naryo"

func main() {
	ctx := context.Background()

	resolveConfigPath()

	configManagers, err := coresetup.SetupConfigManagers(ctx, defaultConfigYAML, rootProperty)
	if err != nil {
		logging.ErrorWithCtx(ctx, "failed to set up configuration managers", "error", err)
		os.Exit(1)
	}

	if err := persistencegormsetup.Migrate(ctx, defaultConfigYAML, rootProperty); err != nil {
		logging.ErrorWithCtx(ctx, "failed to apply database migrations", "error", err)
		os.Exit(1)
	}

	persistenceGormModule, err := persistencegormsetup.Setup(ctx, defaultConfigYAML, rootProperty)
	if err != nil {
		logging.ErrorWithCtx(ctx, "failed to set up persistence-gorm module", "error", err)
		os.Exit(1)
	}

	// broadcaster-http's Setup depends on the HttpClientConfigManager built above.
	broadcasterHttpModule, err := broadcasterhttpsetup.Setup(ctx, configManagers.HttpClientConfigManager, configManagers.BroadcasterConfigConfigManager)
	if err != nil {
		logging.ErrorWithCtx(ctx, "failed to set up broadcaster-http module", "error", err)
		os.Exit(1)
	}

	coreModule, err := coresetup.Setup(
		configManagers,
		persistenceGormModule.EventStores,
		persistenceGormModule.LatestBlockStores,
		[]broadcast.Producer{broadcasterHttpModule.HttpBroadcasterProducer},
	)
	if err != nil {
		logging.ErrorWithCtx(ctx, "failed to set up core module", "error", err)
		os.Exit(1)
	}

	if err := coreModule.Bootstrapper.Start(ctx); err != nil {
		logging.ErrorWithCtx(ctx, "failed to start bootstrapper", "error", err)
		os.Exit(1)
	}

	serverProperties, err := coreconfig.LoadConfig[config.EnvironmentProperties](ctx, defaultConfigYAML, rootProperty)
	if err != nil {
		logging.ErrorWithCtx(ctx, "failed to load server configuration", "error", err)
		os.Exit(1)
	}

	logging.InfoWithCtx(ctx, "server started")
	runHTTPServer(ctx, serverProperties.Server.Port)
}

// resolveConfigPath reads the "-config" flag and, if given, exports it as CONFIG_PATH so every
// LoadConfig call below (one per module's Setup function) picks up the same override file. The
// flag is registered and parsed exactly once, here, since LoadConfig itself no longer touches
// the CLI — it only ever reads CONFIG_PATH from the environment.
func resolveConfigPath() {
	configPath := flag.String("config", "", "Comma-separated list of config files")
	flag.Parse()

	if *configPath != "" {
		_ = os.Setenv("CONFIG_PATH", *configPath)
	}
}

// newHealthMux builds the server's routes. Split out from runHTTPServer so it can be exercised
// directly (via httptest) without binding a real port.
func newHealthMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return mux
}

// runHTTPServer keeps the process up until the user cancels it (Ctrl+C / SIGTERM / SIGKILL):
// ListenAndServe blocks for as long as the server is accepting connections, which is exactly the
// "concurrent flow without final" this module needs — there is no code path here that shuts it
// down on its own.
func runHTTPServer(ctx context.Context, port int) {
	addr := fmt.Sprintf(":%d", port)
	logging.InfoWithCtx(ctx, "HTTP server listening", "addr", addr)
	if err := http.ListenAndServe(addr, newHealthMux()); err != nil {
		logging.ErrorWithCtx(ctx, "HTTP server stopped", "error", err)
		os.Exit(1)
	}
}
