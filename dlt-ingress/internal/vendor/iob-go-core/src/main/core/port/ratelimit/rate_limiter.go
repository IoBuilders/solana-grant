package ratelimit

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
)

type RateLimiter struct {
	buckets []*bucket
	now     func() time.Time
}

type OptionConfig func(limiter *RateLimiter)

type BucketOptionConfig func(b *bucket)

func New(optConfs ...OptionConfig) *RateLimiter {
	limiter := &RateLimiter{now: time.Now}
	for _, optConf := range optConfs {
		optConf(limiter)
	}
	return limiter
}

func WithBucket(name string, requestsPerMinute int, bucketOptConfs ...BucketOptionConfig) OptionConfig {
	return func(limiter *RateLimiter) {
		b := newBucket(name, requestsPerMinute)
		for _, bucketOptConf := range bucketOptConfs {
			bucketOptConf(b)
		}
		limiter.buckets = append(limiter.buckets, b)
	}
}

func Methods(methods ...string) BucketOptionConfig {
	return func(b *bucket) {
		set := make(map[string]struct{}, len(methods))
		for _, method := range methods {
			set[method] = struct{}{}
		}
		b.matcher = methodMatcher{methods: set}
	}
}

func MatchAll() BucketOptionConfig {
	return func(b *bucket) {
		b.matcher = methodMatcher{all: true}
	}
}

func withClock(now func() time.Time) OptionConfig {
	return func(limiter *RateLimiter) {
		limiter.now = now
	}
}

func (l *RateLimiter) Allow(ctx context.Context, method string) (allowed bool, retryAfter time.Duration) {
	b := l.bucketFor(method)
	if b == nil {
		return true, 0
	}

	allowed, retryAfter = b.allow(l.now())
	if !allowed {
		logger.DebugWithCtx(ctx, fmt.Sprintf("rate limit bucket %s exhausted for method %s. Capacity in %v", b.name, method, retryAfter))
	}
	return allowed, retryAfter
}

func (l *RateLimiter) NotifyExceeded(ctx context.Context, method string, resetAt *time.Time) {
	b := l.bucketFor(method)
	if b == nil {
		return
	}

	logger.WarnWithCtx(ctx, fmt.Sprintf("node reported rate limit exceeded for method %s. Bucket %s blocked", method, b.name), "resetAt", resetAt)
	b.notifyExceeded(l.now(), resetAt)
}

func (l *RateLimiter) bucketFor(method string) *bucket {
	for _, b := range l.buckets {
		if b.matcher.matches(method) {
			return b
		}
	}
	return nil
}
