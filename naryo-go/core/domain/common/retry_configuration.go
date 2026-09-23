package common

import (
	"time"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

const (
	defaultMaxRetries   = 5
	defaultInitialDelay = time.Second
	defaultMaxDelay     = 30 * time.Second
	defaultMultiplier   = 2.0
)

// RetryConfiguration drives the exponential backoff applied when a Node
// connection drops: delays start at InitialDelay and grow by Multiplier up to
// MaxDelay, for at most MaxRetries attempts.
type RetryConfiguration struct {
	MaxRetries   int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

// DefaultRetryConfiguration returns the standard backoff policy: 5 retries,
// starting at 1s, doubling up to 30s.
func DefaultRetryConfiguration() *RetryConfiguration {
	return &RetryConfiguration{
		MaxRetries:   defaultMaxRetries,
		InitialDelay: defaultInitialDelay,
		MaxDelay:     defaultMaxDelay,
		Multiplier:   defaultMultiplier,
	}
}

func NewRetryConfiguration(maxRetries int, initialDelay, maxDelay time.Duration, multiplier float64) (*RetryConfiguration, error) {
	r := &RetryConfiguration{
		MaxRetries:   maxRetries,
		InitialDelay: initialDelay,
		MaxDelay:     maxDelay,
		Multiplier:   multiplier,
	}
	if err := r.Validate(); err != nil {
		return &RetryConfiguration{}, err
	}
	return r, nil
}

func (r RetryConfiguration) Validate() error {
	if r.MaxRetries < 0 {
		return domainerrors.NewInvalidFieldError("MaxRetries", "RetryConfiguration", "must be >= 0")
	}
	if r.InitialDelay <= 0 {
		return domainerrors.NewInvalidFieldError("InitialDelay", "RetryConfiguration", "must be > 0")
	}
	if r.MaxDelay < r.InitialDelay {
		return domainerrors.NewInvalidFieldError("MaxDelay", "RetryConfiguration", "must be >= InitialDelay")
	}
	if r.Multiplier < 1 {
		return domainerrors.NewInvalidFieldError("Multiplier", "RetryConfiguration", "must be >= 1")
	}
	return nil
}
