package ratelimit

import (
	"errors"
	"fmt"
	"math"
	"time"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
)

var _ coreerror.RetryAfterError = (*ExceededError)(nil)

type ExceededError struct {
	Method     string
	RetryAfter time.Duration
}

func NewExceededError(method string, retryAfter time.Duration) error {
	return &ExceededError{Method: method, RetryAfter: retryAfter}
}

func (e *ExceededError) Error() string {
	return fmt.Sprintf("upstream rate limit exceeded for method %s. Retry after %v", e.Method, e.RetryAfter)
}

func (e *ExceededError) Unwrap() error {
	return coreerror.ErrTooManyRequests
}

func (e *ExceededError) RetryAfterSeconds() int {
	if e.RetryAfter <= 0 {
		return 0
	}
	return int(math.Ceil(e.RetryAfter.Seconds())) // rounded up: a client retrying too early gets rejected again
}

func IsExceededError(err error) bool {
	_, ok := errors.AsType[*ExceededError](err)
	return ok
}
