package coreconfig

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/retry"
)

const defaultSemaphoreWaitTimeout = 30 * time.Second

type JanitorConfig struct {
	Threshold *time.Duration `mapstructure:"threshold"`
	Interval  *time.Duration `mapstructure:"interval"`
}

type RetryConfig struct {
	MaxAttempts *int           `mapstructure:"maxAttempts"`
	Delay       *time.Duration `mapstructure:"delay"`
	MaxDelay    *time.Duration `mapstructure:"maxDelay"`
	Multiplier  *float64       `mapstructure:"multiplier"`
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

func (rc *RetryConfig) ToRetryOptions(
	exclude []error,
	failAttemptFunc func(ctx context.Context, attempt int, err error) error,
) retry.Options {
	options := []retry.OptionConfig{
		retry.WithExclude(exclude),
		retry.WithFailAttemptFunc(failAttemptFunc),
	}
	if rc.MaxAttempts != nil {
		options = append(options, retry.WithMaxAttempts(*rc.MaxAttempts))
	}
	if rc.Delay != nil {
		options = append(options, retry.WithDelay(*rc.Delay))
	}
	if rc.MaxDelay != nil {
		options = append(options, retry.WithMaxDelay(*rc.MaxDelay))
	}
	if rc.Multiplier != nil {
		options = append(options, retry.WithMultiplier(*rc.Multiplier))
	}
	return retry.NewOptions(options...)
}

func (rc *RetryConfig) ToRetryOptionsWithExcludeLogic(excludeLogic func(err error) bool) retry.Options {
	options := []retry.OptionConfig{
		retry.WithExcludeLogic(excludeLogic),
	}
	if rc.MaxAttempts != nil {
		options = append(options, retry.WithMaxAttempts(*rc.MaxAttempts))
	}
	if rc.Delay != nil {
		options = append(options, retry.WithDelay(*rc.Delay))
	}
	if rc.MaxDelay != nil {
		options = append(options, retry.WithMaxDelay(*rc.MaxDelay))
	}
	if rc.Multiplier != nil {
		options = append(options, retry.WithMultiplier(*rc.Multiplier))
	}
	return retry.NewOptions(options...)
}

func GetJanitorInterval(janitorConfig *JanitorConfig, defaultJanitorConfig *JanitorConfig) time.Duration {
	var interval *time.Duration
	if janitorConfig != nil {
		interval = janitorConfig.Interval
	}

	if interval == nil && defaultJanitorConfig != nil {
		interval = defaultJanitorConfig.Interval
	}

	if interval == nil {
		return 30 * time.Second
	}

	return *interval
}

func GetJanitorThreshold(janitorConfig *JanitorConfig, defaultJanitorConfig *JanitorConfig) time.Duration {
	var threshold *time.Duration
	if janitorConfig != nil {
		threshold = janitorConfig.Threshold
	}

	if threshold == nil && defaultJanitorConfig != nil {
		threshold = defaultJanitorConfig.Threshold
	}

	if threshold == nil {
		return 180 * time.Second
	}

	return *threshold
}

func GetEventStoreInterval(eventStoreConfig *EventStoreConfig, defaultEventStoreConfig *EventStoreConfig) time.Duration {
	var interval *time.Duration
	if eventStoreConfig != nil {
		interval = eventStoreConfig.Interval
	}

	if interval == nil && defaultEventStoreConfig != nil {
		interval = defaultEventStoreConfig.Interval
	}

	if interval == nil {
		return 5 * time.Second
	}

	return *interval
}

func GetEventStoreBatchSize(eventStoreConfig *EventStoreConfig, defaultEventStoreConfig *EventStoreConfig) int {
	var batchSize *int
	if eventStoreConfig != nil {
		batchSize = eventStoreConfig.BatchSize
	}

	if batchSize == nil && defaultEventStoreConfig != nil {
		batchSize = defaultEventStoreConfig.BatchSize
	}

	if batchSize == nil {
		return 100
	}

	return *batchSize
}

func GetSemaphoreWaitTimeout(eventStoreConfig *EventStoreConfig, defaultEventStoreConfig *EventStoreConfig) time.Duration {
	var timeout *time.Duration
	if eventStoreConfig != nil {
		timeout = eventStoreConfig.SemaphoreWaitTimeout
	}

	if timeout == nil && defaultEventStoreConfig != nil {
		timeout = defaultEventStoreConfig.SemaphoreWaitTimeout
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
func MustValidateSemaphoreTimeout(bc string, eventStoreConfig *EventStoreConfig, defaultEventStoreConfig *EventStoreConfig, janitorConfig *JanitorConfig, defaultJanitorConfig *JanitorConfig) {
	timeout := GetSemaphoreWaitTimeout(eventStoreConfig, defaultEventStoreConfig)
	threshold := GetJanitorThreshold(janitorConfig, defaultJanitorConfig)
	if timeout <= 0 {
		panic(fmt.Sprintf("[%s] eventStore.semaphoreWaitTimeout must be > 0, got %s", bc, timeout))
	}
	if timeout >= threshold {
		panic(fmt.Sprintf("[%s] eventStore.semaphoreWaitTimeout (%s) must be < janitor.threshold (%s)", bc, timeout, threshold))
	}
}

func AmountDecodeHook() any {
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
