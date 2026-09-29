package retry

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
)

type CustomRetryer struct {
}

func NewCustomRetryer() CustomRetryer {
	return CustomRetryer{}
}

func (r CustomRetryer) Execute(ctx context.Context, opts Options, fn func(ctx context.Context) (any, error)) (any, error) {
	if err := validateOpts(opts); err != nil {
		return nil, err
	}
	var err error
	var result any
	var currentDelay = opts.delay

	for i := 0; i < opts.maxAttempts; i++ {

		if result, err = fn(ctx); err == nil || shouldExcludeError(opts, err) {
			if err != nil {
				executeFailAttemptFunction(ctx, opts, err, i+1)
			}
			return result, err // Success or error exclusion
		}

		executeFailAttemptFunction(ctx, opts, err, i+1)

		// Last attempt
		if i == opts.maxAttempts-1 {
			break
		}

		logger.DebugWithCtx(ctx, fmt.Sprintf("attempt %d failed. Error was: %s. Retrying in %v seconds...", i+1, err, currentDelay))

		select {
		case <-time.After(currentDelay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}

		// Apply multiplier for next attempt
		currentDelay = time.Duration(float64(currentDelay) * opts.multiplier)

		if opts.maxDelay > 0 && currentDelay > opts.maxDelay {
			currentDelay = opts.maxDelay
		}
	}
	logger.ErrorWithCtx(ctx, fmt.Sprintf("last attempt failed after %d attempts", opts.maxAttempts), "error", err)
	return nil, err
}

func shouldExcludeError(opts Options, err error) bool {
	for _, e := range opts.exclude {
		if errors.Is(err, e) {
			return true
		}
	}
	if opts.excludeLogic != nil {
		return opts.excludeLogic(err)
	}
	return false
}

func executeFailAttemptFunction(ctx context.Context, opts Options, err error, attemptNumber int) {
	if opts.failAttemptFunc != nil {
		if failedAttemptFuncErr := opts.failAttemptFunc(ctx, attemptNumber, err); failedAttemptFuncErr != nil {
			logger.ErrorWithCtx(ctx, fmt.Sprintf("error executing failed attempt function after failed attempt: %v. Attempt error: %v", failedAttemptFuncErr, err))
		}
	}
}

func validateOpts(opts Options) error {
	if opts.maxAttempts <= 0 {
		return fmt.Errorf("invalid maxAttempts %d. Should be 1 or greater", opts.maxAttempts)
	}
	if opts.delay < time.Duration(1)*time.Second {
		return fmt.Errorf("invalid delay %d. Should be 1 second or greater", opts.delay)
	}
	if opts.maxDelay < time.Duration(1)*time.Second {
		return fmt.Errorf("invalid maxDelay %d. Should be 1 second or greater", opts.maxDelay)
	}
	if opts.multiplier < 1 {
		return fmt.Errorf("invalid multiplier %v. Should be 1.0 or greater", opts.multiplier)
	}
	return nil
}
