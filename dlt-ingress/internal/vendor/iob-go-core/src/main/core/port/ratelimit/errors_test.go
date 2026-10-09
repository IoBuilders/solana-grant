package ratelimit

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
)

func TestExceededError_Message(t *testing.T) {
	err := NewExceededError("eth_call", 5*time.Second)

	assert.Equal(t, "upstream rate limit exceeded for method eth_call. Retry after 5s", err.Error())
}

func TestExceededError_Exposes_Method_And_RetryAfter(t *testing.T) {
	err := fmt.Errorf("sending transaction: %w", NewExceededError("eth_sendRawTransaction", 30*time.Second))

	exceeded, ok := errors.AsType[*ExceededError](err)

	assert.True(t, ok)
	assert.Equal(t, "eth_sendRawTransaction", exceeded.Method)
	assert.Equal(t, 30*time.Second, exceeded.RetryAfter) // what a retryable listener schedules its next attempt with
}

func TestIsExceededError(t *testing.T) {
	assert.True(t, IsExceededError(NewExceededError("eth_call", time.Second)))
	assert.True(t, IsExceededError(fmt.Errorf("sending transaction: %w", NewExceededError("eth_call", time.Second))))
	assert.False(t, IsExceededError(errors.New("some other error")))
	assert.False(t, IsExceededError(nil))
}

func TestExceededError_IsTooManyRequests(t *testing.T) {
	err := fmt.Errorf("reading allowance: %w", NewExceededError("eth_call", 30*time.Second))

	// what makes the HTTP layer answer 429 without knowing about rate limiting
	assert.True(t, errors.Is(err, coreerror.ErrTooManyRequests))
	assert.False(t, errors.Is(errors.New("boom"), coreerror.ErrTooManyRequests))
}

func TestExceededError_RetryAfterSeconds(t *testing.T) {
	tests := []struct {
		name       string
		retryAfter time.Duration
		want       int
	}{
		{"exact seconds", 30 * time.Second, 30},
		{"rounds up", 600 * time.Millisecond, 1},
		{"rounds up a fraction", 2500 * time.Millisecond, 3},
		{"zero", 0, 0},
		{"negative", -1 * time.Second, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &ExceededError{Method: "eth_call", RetryAfter: tt.retryAfter}
			assert.Equal(t, tt.want, err.RetryAfterSeconds())
		})
	}
}
