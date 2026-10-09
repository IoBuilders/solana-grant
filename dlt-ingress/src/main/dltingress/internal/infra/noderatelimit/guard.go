package noderatelimit

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"dlt-ingress/src/main/config"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/ratelimit"
)

// Guard puts a network's rate limiter in front of the node RPC calls, and corrects it whenever the node
// itself answers with a 429.
type Guard struct {
	limiter                *ratelimit.RateLimiter
	retryAfterHeader       string
	retryAfterHeaderFormat config.RetryAfterHeaderFormat
	retryAfterMultiplier   float64
}

func NewGuard(limiter *ratelimit.RateLimiter, rateLimitConfig *config.RateLimitConfig) *Guard {
	guard := &Guard{limiter: limiter, retryAfterMultiplier: 1}
	if rateLimitConfig == nil {
		return guard
	}
	guard.retryAfterHeader = rateLimitConfig.RetryAfterHeader
	guard.retryAfterHeaderFormat = rateLimitConfig.RetryAfterHeaderFormat
	if rateLimitConfig.RetryAfterMultiplier > 1 {
		guard.retryAfterMultiplier = rateLimitConfig.RetryAfterMultiplier
	}
	return guard
}

// Call rejects call up front with a ratelimit.ExceededError when method's bucket has no capacity
// or the node answers with a 429.
func Call[T any](ctx context.Context, g *Guard, method string, call func() (T, error)) (T, error) {
	var zero T
	if allowed, retryAfter := g.limiter.Allow(ctx, method); !allowed {
		return zero, ratelimit.NewExceededError(method, retryAfter)
	}

	result, err := call()
	now := time.Now()
	if resetAt, is429 := g.parse429(err, now); is429 {
		logger.WarnWithCtx(ctx, "node answered with 429", "method", method, "error", err)
		g.limiter.NotifyExceeded(ctx, method, resetAt)
		var retryAfter time.Duration
		if resetAt != nil {
			retryAfter = resetAt.Sub(now)
		}
		return zero, ratelimit.NewExceededError(method, retryAfter)
	}
	return result, err
}

func (g *Guard) parse429(err error, now time.Time) (resetAt *time.Time, is429 bool) {
	if tooManyRequestsErr, ok := errors.AsType[*TooManyRequestsError](err); ok {
		if g.retryAfterHeader == "" {
			return nil, true
		}
		return g.parse429RetryAfterHeader(tooManyRequestsErr, now), true
	}

	return nil, false
}

func (g *Guard) parse429RetryAfterHeader(tooManyRequestsErr *TooManyRequestsError, now time.Time) *time.Time {
	headerValue := strings.TrimSpace(tooManyRequestsErr.Header.Get(g.retryAfterHeader))
	if headerValue == "" {
		return nil
	}

	var resetAt time.Time
	switch g.retryAfterHeaderFormat {
	case config.RetryAfterHeaderFormatSeconds:
		seconds, err := strconv.ParseFloat(headerValue, 64)
		if err != nil || seconds <= 0 {
			return nil
		}
		resetAt = now.Add(time.Duration(seconds * float64(time.Second)))
	case config.RetryAfterHeaderFormatUnixTimestamp:
		timestamp, err := strconv.ParseInt(headerValue, 10, 64)
		if err != nil {
			return nil
		}
		resetAt = time.Unix(timestamp, 0)
	default:
		return nil
	}

	if resetAt.Before(now) {
		return nil
	}
	// The node's reset instant is stretched by a safety margin, so the retry lands after its window really resets
	resetAt = now.Add(time.Duration(float64(resetAt.Sub(now)) * g.retryAfterMultiplier))
	return &resetAt
}
