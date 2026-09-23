package config

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"dario.cat/mergo"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/retry"
)

//go:embed application.yml

var applicationYml string

type Config struct {
	Application struct {
		Name string `mapstructure:"name"`
	} `mapstructure:"application"`

	LoggingConfig struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"` // "text|json"
	} `mapstructure:"logging"`

	Server struct {
		Port                string   `mapstructure:"port"`
		Hostname            string   `mapstructure:"hostname"`
		BaseUrl             string   `mapstructure:"baseUrl"`
		CallbackInterval    string   `mapstructure:"callbackInterval"`
		CallbackMaxAttempts uint     `mapstructure:"callbackMaxAttempts"`
		AllowedOrigins      []string `mapstructure:"allowedOrigins"`
		Health              struct {
			Check struct {
				Path string `mapstructure:"path"`
			} `mapstructure:"check"`
			Status struct {
				Path string `mapstructure:"path"`
			} `mapstructure:"status"`
		} `mapstructure:"health"`
		DefaultRequestTimeout time.Duration `mapstructure:"defaultRequestTimeout"`
		ReadHeaderTimeout     time.Duration `mapstructure:"readHeaderTimeout"`
		ReadTimeout           time.Duration `mapstructure:"readTimeout"`
		WriteTimeout          time.Duration `mapstructure:"writeTimeout"`
		IdleTimeout           time.Duration `mapstructure:"idleTimeout"`
	} `mapstructure:"server"`

	Database struct {
		DltIngress *DatabaseConfig `mapstructure:"dltingress"`
	} `mapstructure:"database"`

	Auth struct {
		Jwt struct {
			Sources []JwtSource `mapstructure:"sources"`
		} `mapstructure:"jwt"`
	} `mapstructure:"auth"`

	RetryableListeners struct {
		Default    *RetryConfig               `mapstructure:"default"`
		DltIngress *BcRetryableListenerConfig `mapstructure:"dltingress"`
	} `mapstructure:"retryableListeners"`

	ListenerConfig struct {
		Defaults   ListenerConfigDefaults `mapstructure:"defaults"`
		DltIngress *BcListenerConfig      `mapstructure:"dltingress"`
	} `mapstructure:"listenerConfig"`

	Janitor struct {
		Default    *JanitorConfig `mapstructure:"default"`
		DltIngress *JanitorConfig `mapstructure:"dltingress"`
	} `mapstructure:"janitor"`

	EventStore struct {
		Default    *EventStoreConfig `mapstructure:"default"`
		DltIngress *EventStoreConfig `mapstructure:"dltingress"`
	} `mapstructure:"eventStore"`

	Otel struct {
		Service  *ServiceOtel  `mapstructure:"service"`
		Resource *ResourceOtel `mapstructure:"resource"`
		Global   *GlobalOtel   `mapstructure:"global"`
		Logs     *LogsOtel     `mapstructure:"logs"`
		Traces   *TracesOtel   `mapstructure:"traces"`
		Metrics  *MetricsOtel  `mapstructure:"metrics"`
	} `mapstructure:"otel"`

	Cache struct {
		Caches map[string]time.Duration `mapstructure:"caches"`
	} `mapstructure:"cache"`

	DltIngress DltIngress `mapstructure:"dltingress"`
}

type JwtSource struct {
	Issuer    string `mapstructure:"issuer"`
	Uri       string `mapstructure:"uri"`
	PublicKey string `mapstructure:"publicKey"`
}

type DltIngress struct {
	Networks               []*NetworkConfig `mapstructure:"networks"`
	Custody                CustodyConfig    `mapstructure:"custody"`
	TxBoundedBlockingQueue struct {
		ExpirationTime       time.Duration `mapstructure:"expirationTime"`
		ExpiredCleanInterval time.Duration `mapstructure:"expiredCleanInterval"`
		Retry                RetryConfig   `mapstructure:"retry"`
	} `mapstructure:"txBoundedBlockingQueue"`
	// LockTimeoutRetryableListeners configures the retry policy shared by every listener registered via
	// ListenerRegistrar.RegisterLockTimeoutRetryable — any listener whose DB write can hit a lock timeout
	// (SQLSTATE 55P03), not just order/DLT-nonce listeners.
	LockTimeoutRetryableListeners struct {
		Retry RetryConfig `mapstructure:"retry"`
	} `mapstructure:"lockTimeoutRetryableListeners"`
}

type NetworkConfig struct {
	Id                   string        `mapstructure:"id"`
	Url                  string        `mapstructure:"url"`
	Dlt                  string        `mapstructure:"dlt"`
	ChainId              amount.Amount `mapstructure:"chainId"`
	TransactionType      uint          `mapstructure:"transactionType"`
	GasLimit             amount.Amount `mapstructure:"gasLimit"`
	GasLimitMultiplier   amount.Amount `mapstructure:"gasLimitMultiplier"`
	GasPrice             amount.Amount `mapstructure:"gasPrice"`
	MaxPriorityFeePerGas amount.Amount `mapstructure:"maxPriorityFeePerGas"`
	FeeMultiplier        amount.Amount `mapstructure:"feeMultiplier"`
	MaxTxPoolSize        int           `mapstructure:"maxTxPoolSize"`
	MaxCuPrice           amount.Amount `mapstructure:"maxCuPrice"`
	Contracts            struct {
		Factory               string `mapstructure:"factory"`
		BusinessLogicResolver string `mapstructure:"businessLogicResolver"`
		MarketFactory         string `mapstructure:"marketFactory"`
		SettlementHubFactory  string `mapstructure:"settlementHubFactory"`
		ExternalListFactory   string `mapstructure:"externalListFactory"`
	} `mapstructure:"contracts"`
}

type KMSTagConfig struct {
	Key   string `mapstructure:"key" json:"key"`
	Value string `mapstructure:"value" json:"value"`
}

type CustodyConfig struct {
	Provider string `mapstructure:"provider"`
	Dfns     struct {
		BaseUrl      string `mapstructure:"baseUrl"`
		AuthToken    string `mapstructure:"authToken"`
		CredentialID string `mapstructure:"credentialId"`
		PrivateKey   string `mapstructure:"privateKey"`
	} `mapstructure:"dfns"`
	Kms struct {
		Region      string         `mapstructure:"region"`
		AccessKey   string         `mapstructure:"accessKey"`
		SecretKey   string         `mapstructure:"secretKey"`
		Endpoint    string         `mapstructure:"endpoint"`
		Tags        []KMSTagConfig `mapstructure:"tags"`
		AliasPrefix string         `mapstructure:"aliasPrefix"`
	} `mapstructure:"kms"`
}

// Otel
type (
	ServiceOtel struct {
		Name string `mapstructure:"name" otel_env_name:"OTEL_SERVICE_NAME"`
	}
	ResourceOtel struct {
		Attributes string `mapstructure:"attributes" otel_env_name:"OTEL_RESOURCE_ATTRIBUTES"`
	}
	GlobalOtel struct {
		Endpoint string `mapstructure:"endpoint" otel_env_name:"OTEL_EXPORTER_OTLP_ENDPOINT"`
		Protocol string `mapstructure:"protocol" otel_env_name:"OTEL_EXPORTER_OTLP_PROTOCOL"`
		Headers  string `mapstructure:"headers" otel_env_name:"OTEL_EXPORTER_OTLP_HEADERS"`
		Insecure string `mapstructure:"insecure" otel_env_name:"OTEL_EXPORTER_OTLP_INSECURE"`
	}
	LogsOtel struct {
		Exporter string `mapstructure:"exporter" otel_env_name:"OTEL_LOGS_EXPORTER"`
		Endpoint string `mapstructure:"endpoint" otel_env_name:"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT"`
		Protocol string `mapstructure:"protocol" otel_env_name:"OTEL_EXPORTER_OTLP_LOGS_PROTOCOL"`
		Headers  string `mapstructure:"headers" otel_env_name:"OTEL_EXPORTER_OTLP_LOGS_HEADERS"`
		Insecure string `mapstructure:"insecure" otel_env_name:"OTEL_EXPORTER_OTLP_LOGS_INSECURE"`
	}
	TracesOtel struct {
		Exporter string `mapstructure:"exporter" otel_env_name:"OTEL_TRACES_EXPORTER"`
		Endpoint string `mapstructure:"endpoint" otel_env_name:"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"`
		Protocol string `mapstructure:"protocol" otel_env_name:"OTEL_EXPORTER_OTLP_TRACES_PROTOCOL"`
		Headers  string `mapstructure:"headers" otel_env_name:"OTEL_EXPORTER_OTLP_TRACES_HEADERS"`
		Insecure string `mapstructure:"insecure" otel_env_name:"OTEL_EXPORTER_OTLP_TRACES_INSECURE"`
	}
	MetricsOtel struct {
		Exporter       string `mapstructure:"exporter" otel_env_name:"OTEL_METRICS_EXPORTER"`
		Endpoint       string `mapstructure:"endpoint" otel_env_name:"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT"`
		Protocol       string `mapstructure:"protocol" otel_env_name:"OTEL_EXPORTER_OTLP_METRICS_PROTOCOL"`
		Headers        string `mapstructure:"headers" otel_env_name:"OTEL_EXPORTER_OTLP_METRICS_HEADERS"`
		Insecure       string `mapstructure:"insecure" otel_env_name:"OTEL_EXPORTER_OTLP_METRICS_INSECURE"`
		Interval       uint   `mapstructure:"interval" otel_env_name:"OTEL_METRIC_EXPORT_INTERVAL"`
		PrometheusHost string `mapstructure:"prometheusHost" otel_env_name:"OTEL_EXPORTER_PROMETHEUS_HOST"`
		PrometheusPort string `mapstructure:"prometheusPort" otel_env_name:"OTEL_EXPORTER_PROMETHEUS_PORT"`
	}
)

type JanitorConfig struct {
	Threshold *time.Duration `mapstructure:"threshold"`
	Interval  *time.Duration `mapstructure:"interval"`
}

type EventFilter struct {
	EventName       string   `mapstructure:"eventName"`
	ContractAddress string   `mapstructure:"contractAddress"`
	Urls            []string `mapstructure:"urls"`
}

type TransactionFilter struct {
	ContractAddress string   `mapstructure:"contractAddress"`
	Urls            []string `mapstructure:"urls"`
}

type RetryConfig struct {
	MaxAttempts *int           `mapstructure:"maxAttempts"`
	Delay       *time.Duration `mapstructure:"delay"`
	MaxDelay    *time.Duration `mapstructure:"maxDelay"`
	Multiplier  *float64       `mapstructure:"multiplier"`
}

func (rc *RetryConfig) ToRetryOptions(
	exclude []error,
	failAttemptFunc func(ctx context.Context, attempt int, err error) error,
) retry.Options {
	return retry.NewOptions(
		retry.WithMaxAttempts(*rc.MaxAttempts),
		retry.WithDelay(*rc.Delay),
		retry.WithMaxDelay(*rc.MaxDelay),
		retry.WithMultiplier(*rc.Multiplier),
		retry.WithExclude(exclude),
		retry.WithFailAttemptFunc(failAttemptFunc),
	)
}

func (rc *RetryConfig) ToRetryOptionsWithExcludeLogic(excludeLogic func(err error) bool) retry.Options {
	return retry.NewOptions(
		retry.WithMaxAttempts(*rc.MaxAttempts),
		retry.WithDelay(*rc.Delay),
		retry.WithMaxDelay(*rc.MaxDelay),
		retry.WithMultiplier(*rc.Multiplier),
		retry.WithExcludeLogic(excludeLogic),
	)
}

type BcRetryableListenerConfig struct {
	Default   *RetryConfig            `mapstructure:"default"`
	Listeners map[string]*RetryConfig `mapstructure:"listeners"`
}

// BcListenerConfig holds per-listener concurrency and execution timeout overrides for a bounded context.
// Concurrency resolution: Listeners[name] → ConcurrencyDefault → global ListenerConfig.Defaults.Concurrency → event.DefaultListenerMaxConcurrency.
// Timeout resolution: ListenerTimeouts[name] → ExecutionTimeoutDefault → global ListenerConfig.Defaults.ExecutionTimeout → 0 (no timeout).
type BcListenerConfig struct {
	ConcurrencyDefault      *int                      `mapstructure:"concurrencyDefault"`
	ExecutionTimeoutDefault *time.Duration            `mapstructure:"executionTimeoutDefault"`
	Listeners               map[string]*int           `mapstructure:"listeners"`
	ListenerTimeouts        map[string]*time.Duration `mapstructure:"listenerTimeouts"`
}

// ListenerConfigDefaults groups the listenerConfig defaults. Concurrency/ExecutionTimeout are the
// lowest-priority fallback used by any listener with no per-listener or BC config. LockTimeoutRetryable
// is the opposite: a highest-priority override applied to every listener registered via
// ListenerRegistrar.RegisterLockTimeoutRetryable (any listener whose DB write can hit a lock timeout,
// not just order/DLT-nonce listeners), so it doesn't need to be repeated under each bounded context's
// `listeners` map.
type ListenerConfigDefaults struct {
	Concurrency          *int                                   `mapstructure:"concurrency"`
	ExecutionTimeout     *time.Duration                         `mapstructure:"executionTimeout"`
	LockTimeoutRetryable ListenerConfigLockTimeoutRetryDefaults `mapstructure:"lockTimeoutRetryable"`
}

type ListenerConfigLockTimeoutRetryDefaults struct {
	Concurrency      *int           `mapstructure:"concurrency"`
	ExecutionTimeout *time.Duration `mapstructure:"executionTimeout"`
}

type EventStoreConfig struct {
	Interval             *time.Duration `mapstructure:"interval"`
	BatchSize            *int           `mapstructure:"batchSize"`
	SemaphoreWaitTimeout *time.Duration `mapstructure:"semaphoreWaitTimeout"`
}

type DatabaseConfig struct {
	Url                             string        `mapstructure:"url"`
	MaxOpenConnections              int           `mapstructure:"maxOpenConnections"`
	MaxIdleConnections              int           `mapstructure:"maxIdleConnections"`
	ConnMaxLifetime                 time.Duration `mapstructure:"connMaxLifetime"`
	ConnMaxIdleTime                 time.Duration `mapstructure:"connMaxIdleTime"`
	ConnectTimeout                  time.Duration `mapstructure:"connectTimeout"`
	StatementTimeout                time.Duration `mapstructure:"statementTimeout"`
	IdleInTransactionSessionTimeout time.Duration `mapstructure:"idleInTransactionSessionTimeout"`
	LockTimeout                     time.Duration `mapstructure:"lockTimeout"`
}

var AppConfig *Config

func LoadConfig(ctx context.Context) error {
	configFilesFlag := flag.String("config", "", "Comma-separated list of config files")
	flag.Parse()

	baseContent := replaceEnvVariables(applicationYml)
	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(bytes.NewBufferString(baseContent)); err != nil {
		return fmt.Errorf("error reading embedded config: %w", err)
	}

	var baseConfig Config
	if err := v.Unmarshal(&baseConfig, viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToWeakSliceHookFunc(","),
			AmountDecodeHook(),
		))); err != nil {
		return fmt.Errorf("error unmarshalling embedded config: %w", err)
	}
	logger.InfoWithCtx(ctx, "Loaded embedded application.yml")

	if *configFilesFlag == "" {
		*configFilesFlag = os.Getenv("CONFIG_PATH")
	}

	if *configFilesFlag != "" {
		files := splitAndTrim(*configFilesFlag, ",")
		for _, file := range files {
			if !fileExists(file) {
				return fmt.Errorf("config file not found: %s", file)
			}
			logger.DebugWithCtx(ctx, "[CONFIG] merging config file", "filename", file)

			fileContent, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("error reading config file %s: %w", file, err)
			}

			overrideContent := replaceEnvVariables(string(fileContent))
			tmpV := viper.New()

			ext := filepath.Ext(file)
			switch ext {
			case ".yaml", ".yml":
				tmpV.SetConfigType("yaml")
			case ".json":
				tmpV.SetConfigType("json")
			case ".toml":
				tmpV.SetConfigType("toml")
			default:
				return fmt.Errorf("unsupported config file extension: %s", ext)
			}

			if err := tmpV.ReadConfig(bytes.NewBufferString(overrideContent)); err != nil {
				return fmt.Errorf("error reading config file %s: %w", file, err)
			}

			var overrideConfig Config
			if err := tmpV.Unmarshal(&overrideConfig); err != nil {
				return fmt.Errorf("error unmarshalling override config: %w", err)
			}

			// Deep merge override into base
			if err := mergo.Merge(&baseConfig, overrideConfig, mergo.WithOverride); err != nil {
				return fmt.Errorf("error merging config: %w", err)
			}
			logger.InfoWithCtx(ctx, "[CONFIG] Successfully merged config file", "filename", file)
		}
	}

	baseConfig.Server.BaseUrl = "http://" + baseConfig.Server.Hostname + ":" + baseConfig.Server.Port

	AppConfig = &baseConfig

	if jsonBytes, err := json.MarshalIndent(baseConfig, "", "  "); err == nil {
		logger.DebugWithCtx(ctx, "[CONFIG] Final merged configuration", "config", string(jsonBytes))
	}

	if err := loadOtelEnvVars(baseConfig); err != nil {
		return err
	}

	return nil
}

func loadOtelEnvVars(baseConfig Config) error {
	sections := []struct {
		cfg  any
		name string
	}{
		{baseConfig.Otel.Service, "service"},
		{baseConfig.Otel.Logs, "logs"},
		{baseConfig.Otel.Resource, "resource"},
		{baseConfig.Otel.Traces, "otel traces"},
		{baseConfig.Otel.Metrics, "metrics"},
		{baseConfig.Otel.Global, "global config"},
	}

	for _, s := range sections {
		if s.cfg != nil {
			if err := loadEnvVars(s.cfg, "otel_env_name"); err != nil {
				return fmt.Errorf("[OTEL] Failed to load otel_env_name variable '%s' %v", s.name, err)
			}
		}
	}
	return nil
}

func loadEnvVars(config any, tagName string) error {
	v := reflect.ValueOf(config)
	if !v.IsValid() {
		return fmt.Errorf("config is nil/invalid")
	}
	skip := v.Kind() != reflect.Pointer || v.IsNil() || v.Elem().Kind() != reflect.Struct

	if skip {
		return nil
	}

	v = v.Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		if f.PkgPath != "" {
			continue
		}

		key := f.Tag.Get(tagName)
		if key == "" || key == "-" {
			continue
		}

		fv := v.Field(i)

		for fv.Kind() == reflect.Pointer {
			if fv.IsNil() {
				fv = reflect.Value{} // mark invalid
				break
			}
			fv = fv.Elem()
		}
		if !fv.IsValid() {
			continue
		}

		val, err := valueToString(fv)
		if err != nil {
			return fmt.Errorf("field %s: %w", f.Name, err)
		}

		if err := os.Setenv(key, val); err != nil {
			return fmt.Errorf("setenv %q from field %s: %w", key, f.Name, err)
		}
	}

	return nil
}

func valueToString(v reflect.Value) (string, error) {
	switch v.Kind() {
	case reflect.String:
		return v.String(), nil
	case reflect.Bool:
		if v.Bool() {
			return "true", nil
		}
		return "false", nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64), nil
	default:
		return "", fmt.Errorf("unsupported kind %s", v.Kind())
	}
}

func (c *Config) GetJanitorInterval(bc string) time.Duration {
	var interval *time.Duration

	switch bc {
	case "dltingress":
		if c.Janitor.DltIngress != nil {
			interval = c.Janitor.DltIngress.Interval
		}
	}

	if interval == nil && c.Janitor.Default != nil {
		interval = c.Janitor.Default.Interval
	}

	if interval == nil {
		return 30 * time.Second
	}

	return *interval
}

func (c *Config) GetJanitorThreshold(bc string) time.Duration {
	var threshold *time.Duration

	switch bc {
	case "dltingress":
		if c.Janitor.DltIngress != nil {
			threshold = c.Janitor.DltIngress.Threshold
		}
	}

	if threshold == nil && c.Janitor.Default != nil {
		threshold = c.Janitor.Default.Threshold
	}

	if threshold == nil {
		return 180 * time.Second
	}

	return *threshold
}

var envPattern = regexp.MustCompile(`\$\{(\w+)(?::([^}]*))?\}`)

// replaceEnvVariables replaces all placeholders ${VAR:default} in the entire document
// with the corresponding environment variable values or defaults if not set.
func replaceEnvVariables(content string) string {
	return envPattern.ReplaceAllStringFunc(content, func(match string) string {
		submatches := envPattern.FindStringSubmatch(match)
		if len(submatches) != 3 {
			return match
		}
		envVar := submatches[1]
		defaultVal := submatches[2]
		if val, exists := os.LookupEnv(envVar); exists {
			return val
		}
		return defaultVal
	})
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func splitAndTrim(s, sep string) []string {
	raw := bytes.Split([]byte(s), []byte(sep))
	var result []string
	for _, part := range raw {
		trimmed := string(bytes.TrimSpace(part))
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func checkEventFiltersUniqueness(filters []EventFilter) error {
	eventNames := make(map[string]bool)
	for _, filter := range filters {
		if _, exists := eventNames[filter.EventName]; exists {
			return fmt.Errorf("event with name %s is duplicated", filter.EventName)
		}
		eventNames[filter.EventName] = true
	}
	return nil
}

func ExtractBoundedContext(url string) string {
	const prefix = "/api/v1/internal/"
	idx := strings.Index(url, prefix)
	if idx < 0 {
		return ""
	}
	rest := url[idx+len(prefix):]
	if end := strings.Index(rest, "/"); end > 0 {
		return rest[:end]
	}
	return rest
}

func (c *Config) GetEventStoreInterval(bc string) time.Duration {
	var interval *time.Duration

	switch bc {
	case "dltingress":
		if c.EventStore.DltIngress != nil {
			interval = c.EventStore.DltIngress.Interval
		}
	}

	if interval == nil && c.EventStore.Default != nil {
		interval = c.EventStore.Default.Interval
	}

	if interval == nil {
		return 5 * time.Second
	}

	return *interval
}

func (c *Config) GetEventStoreBatchSize(bc string) int {
	var batchSize *int

	switch bc {
	case "dltingress":
		if c.EventStore.DltIngress != nil {
			batchSize = c.EventStore.DltIngress.BatchSize
		}
	}

	if batchSize == nil && c.EventStore.Default != nil {
		batchSize = c.EventStore.Default.BatchSize
	}

	if batchSize == nil {
		return 100
	}

	return *batchSize
}

// defaultSemaphoreWaitTimeout mirrors event.DefaultSemaphoreWaitTimeout to avoid
// importing the event package in this package.
const defaultSemaphoreWaitTimeout = 30 * time.Second

func (c *Config) GetSemaphoreWaitTimeout(bc string) time.Duration {
	var timeout *time.Duration

	switch bc {
	case "dltingress":
		if c.EventStore.DltIngress != nil {
			timeout = c.EventStore.DltIngress.SemaphoreWaitTimeout
		}
	}

	if timeout == nil && c.EventStore.Default != nil {
		timeout = c.EventStore.Default.SemaphoreWaitTimeout
	}

	if timeout == nil {
		return defaultSemaphoreWaitTimeout
	}

	return *timeout
}

// MustValidateSemaphoreTimeout panics at startup if the semaphore wait timeout for the
// given bounded context is not in the valid range (0, janitorThreshold). This prevents
// the relay from parking goroutines long enough for the janitor to race and re-deliver
// the same consumer concurrently.
func (c *Config) MustValidateSemaphoreTimeout(bc string) {
	timeout := c.GetSemaphoreWaitTimeout(bc)
	threshold := c.GetJanitorThreshold(bc)
	if timeout <= 0 {
		panic(fmt.Sprintf("[%s] eventStore.semaphoreWaitTimeout must be > 0, got %s", bc, timeout))
	}
	if timeout >= threshold {
		panic(fmt.Sprintf("[%s] eventStore.semaphoreWaitTimeout (%s) must be < janitor.threshold (%s)", bc, timeout, threshold))
	}
}

func (c *Config) GetDltIngressNetwork(networkId string) (*NetworkConfig, error) {
	for _, network := range c.DltIngress.Networks {
		if network.Id == networkId {
			return network, nil
		}
	}
	return nil, domainerrors.NewEntityNotFoundDomainError("Network", networkId)
}

func (c *Config) GetNetworkOrDefault(networkId *string) (*NetworkConfig, error) {
	if networkId != nil {
		return c.GetDltIngressNetwork(*networkId)
	}
	return c.DltIngress.Networks[0], nil
}

func (n *NetworkConfig) CheckDltAccount(dltAccountId string, dlt string) error {
	if n.Dlt != dlt {
		return coreerror.NewConflictDomainError(
			"DLT_ACCOUNT_ID_NOT_VALID_FOR_NETWORK",
			fmt.Sprintf("DLT Account Id %s of type %s is not valid for network %s. Type should be %s", dltAccountId, dlt, n.Id, n.Dlt),
		)
	}
	return nil
}

func (c *Config) GetDefaultDlt() string {
	if len(c.DltIngress.Networks) == 0 {
		return ""
	}
	return c.DltIngress.Networks[0].Dlt
}

func AmountDecodeHook() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		// Check if the target type is our Amount struct
		if t != reflect.TypeOf(amount.Amount{}) {
			return data, nil
		}

		var valStr string
		switch v := data.(type) {
		case float64:
			valStr = strconv.FormatFloat(v, 'f', -1, 64)
		case string:
			valStr = v
		default:
			valStr = fmt.Sprintf("%d", v)
		}

		return amount.NewFromString(valStr)
	}
}
