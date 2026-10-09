package retry

import (
	"context"
	"time"
)

type Retryer interface {
	Execute(ctx context.Context, opts Options, fn func(ctx context.Context) (any, error)) (any, error)
}

type Options struct {
	maxAttempts     int                                                     // number of max attempts
	delay           time.Duration                                           // initial delay
	maxDelay        time.Duration                                           // max delay
	multiplier      float64                                                 // multiplier to multiply delay in each attempt
	failAttemptFunc func(ctx context.Context, attempt int, err error) error // Function to execute every time an attempt has failed
	exclude         []error                                                 // error slice to exclude from retries
	excludeLogic    func(err error) bool                                    // custom logic to exclude errors from retries
}

type OptionConfig func(opts *Options)

// NonRetryableError marks err as terminal: CustomRetryer stops after the
// first attempt regardless of the configured Options.exclude/excludeLogic.
type NonRetryableError struct{ error }

// NewNonRetryableError wraps err so the retryer treats it as a terminal
// failure instead of retrying it.
func NewNonRetryableError(err error) error {
	if err == nil {
		return nil
	}
	return NonRetryableError{err}
}

func (e NonRetryableError) Unwrap() error {
	return e.error
}

func NewOptions(optConfs ...OptionConfig) Options {
	options := Options{
		maxAttempts: 3,
		delay:       time.Duration(1) * time.Second,
		maxDelay:    time.Duration(30) * time.Second,
		multiplier:  2,
	}
	for _, optConf := range optConfs {
		optConf(&options)
	}
	return options
}

func WithMaxAttempts(maxAttempts int) OptionConfig {
	return func(opts *Options) {
		opts.maxAttempts = maxAttempts
	}
}

func WithDelay(delay time.Duration) OptionConfig {
	return func(opts *Options) {
		opts.delay = delay
	}
}

func WithMaxDelay(maxDelay time.Duration) OptionConfig {
	return func(opts *Options) {
		opts.maxDelay = maxDelay
	}
}

func WithMultiplier(multiplier float64) OptionConfig {
	return func(opts *Options) {
		opts.multiplier = multiplier
	}
}

// WithExclude stops retrying on errors matching any of exclude. A coreerror.RetryAfterError is always
// retried regardless.
func WithExclude(exclude []error) OptionConfig {
	return func(opts *Options) {
		opts.exclude = exclude
	}
}

// WithExcludeLogic stops retrying on errors for which excludeLogic returns true. A
// coreerror.RetryAfterError is always retried regardless.
func WithExcludeLogic(excludeLogic func(err error) bool) OptionConfig {
	return func(opts *Options) {
		opts.excludeLogic = excludeLogic
	}
}

func WithFailAttemptFunc(failAttemptFunc func(ctx context.Context, attempt int, err error) error) OptionConfig {
	return func(opts *Options) {
		opts.failAttemptFunc = failAttemptFunc
	}
}
