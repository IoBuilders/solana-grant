//go:build test || integration

package integration

import (
	"bytes"
	"context"
	"dlt-ingress/src/main/dltingress/internal/infra/svm"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/config"
	"dlt-ingress/src/main/dltingress/internal/domain/common"
	custodymocks "dlt-ingress/src/main/dltingress/internal/infra/custody"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api/middleware"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/cache"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/command"
	coreconfig "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/config"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/panicinfo"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/health"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/observability/metrics"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/retry"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/query"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric/noop"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var databaseUrlMap map[string]string
var ctx context.Context
var pgContainer *postgres.PostgresContainer
var serverUrl string
var DltIngress *dltingressconfig.DltIngress

func TestMain(m *testing.M) {
	err := os.Setenv("TZ", "UTC")
	if err != nil {
		panic("Failed to set timezone")
	}
	time.Local = time.UTC
	exitCode := IntegrationTestEnvironment(m)
	os.Exit(exitCode)
}

func HttpCallJSONResponse[T any](method string, path string, payload any, expectedStatus int) (T, api.ErrorResponse, *http.Response, error) {
	var zero T
	var errorResponse api.ErrorResponse

	b, err := json.Marshal(payload)
	if err != nil {
		return zero, errorResponse, nil, fmt.Errorf("marshal payload: %w", err)
	}

	url := APIPath("%s", path)
	resp, err := SendHTTPRequest(method, url, bytes.NewReader(b))
	if err != nil {
		return zero, errorResponse, resp, err
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != expectedStatus {
		body, _ := io.ReadAll(resp.Body)
		json.Unmarshal(body, &errorResponse)
		return zero, errorResponse, resp, fmt.Errorf("unexpected status %d (want %d): %s",
			resp.StatusCode, expectedStatus, strings.TrimSpace(string(body)))
	}

	var out T
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return zero, errorResponse, resp, fmt.Errorf("decode response: %w", err)
	}
	return out, errorResponse, resp, nil
}

func APIPath(format string, args ...any) string {
	return serverUrl + fmt.Sprintf(format, args...)
}

func IntegrationTestEnvironment(m *testing.M) int {
	originalLoc, _ := time.LoadLocation(time.Local.String())
	systemTime := time.Now().In(originalLoc)

	_ = os.Setenv("TZ", "UTC")

	loc, _ := time.LoadLocation("UTC")
	time.Local = loc

	defer func() {
		if r := recover(); r != nil {
			info := panicinfo.NewPanicInfo(r)
			logger.Error(
				fmt.Sprintf("%s %s", panicinfo.PANIC_RECOVERED_TAG, "Recovered from panic"),
				"panic", info.Message,
				"file", info.File,
				"line", info.Line,
				"stack", info.FormatStackTrace())
			os.Exit(1)
		}
	}()

	ctx = context.Background()

	logger.InfoWithCtx(ctx, "System local time", "time", systemTime)
	logger.InfoWithCtx(ctx, "Backend local time", "time", time.Now())

	if pgContainer == nil {
		databaseUrlMap, pgContainer = InitPostgresContainer(ctx)
	}

	dltIngressConfig := &config.Config{}
	dltIngressConfig.Database = &coreconfig.DatabaseConfig{}
	dltIngressConfig.DltIngress.Networks = []*config.NetworkConfig{
		{
			Dlt: "SVM",
		},
	}
	dltIngressConfig.DltIngress.Custody.Provider = "KMS"
	dltIngressConfig.DltIngress.Custody.Kms.Region = "us-east-1"
	dltIngressConfig.DltIngress.TxBoundedBlockingQueue.ExpiredCleanInterval = 1 * time.Hour
	dltIngressConfig.DltIngress.TxBoundedBlockingQueue.ExpirationTime = 2 * time.Hour
	dltIngressConfig.Cache.Caches = map[string]time.Duration{svm.CacheKey: 30 * time.Second}

	metricsRegistry := metrics.NewRegistry()
	coreDependencies := &dltingressconfig.CoreDependencies{
		QueryBus:        query.NewQueryBus(),
		CrossQueryBus:   query.NewCrossQueryBus(),
		CrossCommandBus: command.NewCrossCommandBus([]string{}, metricsRegistry),
		Retryer:         retry.NewCustomRetryer(),
		Tracer:          otel.GetTracerProvider().Tracer("dltingress"),
		Cache:           cache.NewGoCache(dltIngressConfig.Cache.Caches),
		MetricsRegistry: metricsRegistry,
		HealthRegistry:  health.NewRegistry(),
	}
	config.DltIngressConfig = dltIngressConfig

	const dummyBlockchainUrl = "http://localhost:1"
	overrideDltIngressNetwork(dummyBlockchainUrl)
	overrideDatabaseUrls(databaseUrlMap)

	metrics.SetDefaultMeter(noop.Meter{})

	dltIngressModule := dltingressconfig.NewLazyDltIngress(dltIngressConfig, coreDependencies, dltingressconfig.WithCustodyProvider(custodymocks.NewInMemoryProvider()))
	dltIngressModule.Init(ctx)
	DltIngress = dltIngressModule.DltIngress

	engine := ConfigRouter()
	dltingressconfig.RegisterEndpoints(dltIngressModule, engine, func(c *gin.Context) {
		c.Next()
	})

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		logger.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
	addr := listener.Addr().String()
	serverUrl = fmt.Sprintf("http://%s", addr)

	server := &http.Server{
		Handler: engine,
	}

	// Start a server in the background
	go func() {
		logger.Info(fmt.Sprintf("Starting server on %s...", addr))
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Run the test suite
	exitCode := m.Run()

	// Clean shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server failed to shutdown", "error", err)
	}

	return exitCode
}

func InitPostgresContainer(ctx context.Context) (map[string]string, *postgres.PostgresContainer) {
	if os.Getenv("CI") == "true" {
		logger.Info("🚀 Running in GitLab CI")
	} else {
		logger.Info("🏡 Running locally")
	}

	pgContainer, err := postgres.Run(ctx,
		"postgres:14.7",
		postgres.WithDatabase("dltingress"),
		postgres.WithUsername("dltingress"),
		postgres.WithPassword("dltingress"),
		testcontainers.WithCmd(
			"-c", "max_connections=500",
			"-c", "fsync=off",
			"-c", "synchronous_commit=off",
			"-c", "full_page_writes=off",
		),

		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		logger.Error("failed to start postgreSQL container", "error", err)
		os.Exit(1)
	}
	database, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		logger.Error("failed to get connection string", "error", err)
		os.Exit(1)
	}

	logger.Info(fmt.Sprintf("Connection string: %s", database))
	logger.Info("Test DB container is up and table created successfully")

	dbHost, _ := pgContainer.Host(ctx)
	dbPort, _ := pgContainer.MappedPort(ctx, "5432")

	domains := []string{"dltingress"}
	dbConnections := make(map[string]string)
	for _, dbName := range domains {
		dbConnections[dbName] = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbName, dbName, dbHost, dbPort.Port(), dbName)
	}

	return dbConnections, pgContainer
}

func overrideDatabaseUrls(databaseUrlMap map[string]string) {
	config.DltIngressConfig.Database.Url = databaseUrlMap["dltingress"]
}

func overrideDltIngressNetwork(blockchainUrl string) {
	chainId, _ := amount.NewFromString("1003")
	gasLimit, _ := amount.NewFromString("500000000")
	gasLimitMultiplier, _ := amount.NewFromString("1.25")
	gasPrice, _ := amount.NewFromString("0")
	maxPriorityFeePerGas, _ := amount.NewFromString("0")

	network := &config.NetworkConfig{
		Id:                   "default",
		Url:                  blockchainUrl,
		Dlt:                  string(common.EVM),
		ChainId:              *chainId,
		TransactionType:      0,
		GasLimit:             *gasLimit,
		GasLimitMultiplier:   *gasLimitMultiplier,
		GasPrice:             *gasPrice,
		MaxPriorityFeePerGas: *maxPriorityFeePerGas,
		MaxTxPoolSize:        1000,
	}
	network.Contracts = config.DltIngressConfig.DltIngress.Networks[0].Contracts

	config.DltIngressConfig.DltIngress.Networks = []*config.NetworkConfig{network}
}

func SendHTTPRequest(method string, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	return resp, nil
}

func ConfigRouter() *gin.Engine {
	router := gin.New()

	// Configure CORS
	router.Use(
		middleware.RecoveryMiddleware(),
		middleware.TraceHeaderMiddleware,
	)
	router.Group("api/v1")

	return router
}
