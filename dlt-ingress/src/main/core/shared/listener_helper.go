package shared

import (
	"context"
	"strings"
	"time"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/core/utils"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/config"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/retry"
)

// WrapRetryableListener wraps a listener with retry capabilities using granular configuration.
// It allows passing an extra RetryableListenerOption which will be combined with the one from configuration.
func WrapRetryableListener(
	listener event.Listener,
	retryer retry.Retryer,
	bcConfig *coreconfig.BcRetryableListenerConfig,
	extraOpts ...event.RetryableListenerOption,
) func(ctx context.Context, event event.Event) error {
	mergedCfg := mergeConfig(listener, bcConfig)
	validateAndPanic(mergedCfg)

	opts := extraOpts
	if retryOptionConfigs := ToRetryOptionConfigs(mergedCfg); len(retryOptionConfigs) > 0 {
		opts = append(opts, event.WithRetryOptions(retry.NewOptions(retryOptionConfigs...)))
	}

	return event.NewRetryableListener(listener, retryer, opts...).Listen
}

func validateAndPanic(cfg *coreconfig.RetryConfig) {
	if cfg.MaxAttempts != nil && *cfg.MaxAttempts <= 0 {
		panic("MaxAttempts must be greater than 0")
	}

	if cfg.Delay != nil && *cfg.Delay <= 0 {
		panic("Delay must be greater than 0")
	}

	if cfg.MaxDelay != nil && *cfg.MaxDelay <= 0 {
		panic("MaxDelay must be greater than 0")
	}

	if cfg.Multiplier != nil && *cfg.Multiplier <= 1 {
		panic("Multiplier must be greater than 1")
	}
}

func mergeConfig(listener event.Listener, bcConfig *coreconfig.BcRetryableListenerConfig) *coreconfig.RetryConfig {
	merged := &coreconfig.RetryConfig{}

	// 1. App default configuration
	if config.DltIngressConfig != nil && config.DltIngressConfig.RetryableListeners.Default != nil {
		applyConfig(merged, config.DltIngressConfig.RetryableListeners.Default)
	}

	// 2. Bounded Context default configuration (overrides app default)
	if bcConfig != nil && bcConfig.Default != nil {
		applyConfig(merged, bcConfig.Default)
	}

	// 3. Specific listener configuration (overrides BC and app defaults)
	if bcConfig != nil && bcConfig.Listeners != nil {
		if cfg, ok := bcConfig.Listeners[utils.Key(listener)]; ok {
			applyConfig(merged, cfg)
		}
	}

	return merged
}

func applyConfig(target, source *coreconfig.RetryConfig) {
	if source.MaxAttempts != nil {
		target.MaxAttempts = source.MaxAttempts
	}
	if source.Delay != nil {
		target.Delay = source.Delay
	}
	if source.MaxDelay != nil {
		target.MaxDelay = source.MaxDelay
	}
	if source.Multiplier != nil {
		target.Multiplier = source.Multiplier
	}
}

// WrapLockTimeoutRetryableListener wraps a listener with retry logic that retries only on DB lock
// timeouts (SQLSTATE 55P03), using the dltingress.lockTimeoutRetryableListeners retry config for
// maxAttempts, delay, and backoff settings. Suitable for any listener whose write can be blocked by a
// lock, not just order/DLT-nonce listeners.
func WrapLockTimeoutRetryableListener(
	listener event.Listener,
	retryer retry.Retryer,
) func(ctx context.Context, ev event.Event) error {
	optConfs := ToRetryOptionConfigs(&config.DltIngressConfig.DltIngress.LockTimeoutRetryableListeners.Retry)
	optConfs = append(optConfs, retry.WithExcludeLogic(func(err error) bool {
		return !utils.IsLockTimeoutError(err)
	}))
	return event.NewRetryableListener(
		listener,
		retryer,
		event.WithRetryOptions(retry.NewOptions(optConfs...)),
	).Listen
}

// GetListenerConcurrency resolves the goroutine cap for a named listener.
// Resolution: sharedVar → per-listener override → BC default → global default → 0 (event.DefaultListenerMaxConcurrency).
// sharedVar is optional — pass it to force a single value across a whole category of listeners (e.g.
// every lock-timeout-retryable listener via listenerConfig.defaults.lockTimeoutRetryable.concurrency)
// without having to repeat it under every bounded context's `listeners` map; omit it otherwise.
// YAML keys use the package portion of the listener name (before the first dot) to avoid viper's
// dot-as-key-delimiter splitting behaviour, so "ordercreateasset.Listener" looks up "ordercreateasset".
func GetListenerConcurrency(listenerName string, bcConfig *coreconfig.BcListenerConfig, sharedVar ...*int) int {
	if len(sharedVar) > 0 && sharedVar[0] != nil {
		return *sharedVar[0]
	}
	key := strings.SplitN(listenerName, ".", 2)[0]
	if bcConfig != nil {
		if v, ok := bcConfig.Listeners[key]; ok && v != nil {
			return *v
		}
		if bcConfig.ConcurrencyDefault != nil {
			return *bcConfig.ConcurrencyDefault
		}
	}
	if config.DltIngressConfig != nil && config.DltIngressConfig.ListenerConfig.Defaults.Concurrency != nil {
		return *config.DltIngressConfig.ListenerConfig.Defaults.Concurrency
	}
	return 0
}

// GetListenerExecutionTimeout resolves the execution timeout for a named listener.
// Resolution: sharedVar → per-listener override → BC executionTimeoutDefault → global executionTimeoutDefault → 0 (no timeout).
// sharedVar is optional — pass it to force a single value across a whole category of listeners (e.g.
// every lock-timeout-retryable listener via listenerConfig.defaults.lockTimeoutRetryable.executionTimeout);
// omit it otherwise.
// YAML keys use the package portion of the listener name (before the first dot).
func GetListenerExecutionTimeout(listenerName string, bcConfig *coreconfig.BcListenerConfig, sharedVar ...*time.Duration) time.Duration {
	if len(sharedVar) > 0 && sharedVar[0] != nil {
		return *sharedVar[0]
	}
	key := strings.SplitN(listenerName, ".", 2)[0]
	if bcConfig != nil {
		if v, ok := bcConfig.ListenerTimeouts[key]; ok && v != nil {
			return *v
		}
		if bcConfig.ExecutionTimeoutDefault != nil {
			return *bcConfig.ExecutionTimeoutDefault
		}
	}
	if config.DltIngressConfig != nil && config.DltIngressConfig.ListenerConfig.Defaults.ExecutionTimeout != nil {
		return *config.DltIngressConfig.ListenerConfig.Defaults.ExecutionTimeout
	}
	return 0
}

func ToRetryOptionConfigs(cfg *coreconfig.RetryConfig) []retry.OptionConfig {
	var retryOptionConfigs []retry.OptionConfig

	if cfg != nil {
		if cfg.MaxAttempts != nil {
			retryOptionConfigs = append(retryOptionConfigs, retry.WithMaxAttempts(*cfg.MaxAttempts))
		}
		if cfg.Delay != nil {
			retryOptionConfigs = append(retryOptionConfigs, retry.WithDelay(*cfg.Delay))
		}
		if cfg.MaxDelay != nil {
			retryOptionConfigs = append(retryOptionConfigs, retry.WithMaxDelay(*cfg.MaxDelay))
		}
		if cfg.Multiplier != nil {
			retryOptionConfigs = append(retryOptionConfigs, retry.WithMultiplier(*cfg.Multiplier))
		}
	}

	return retryOptionConfigs
}
