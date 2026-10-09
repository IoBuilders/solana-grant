package ratelimit

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type methodMatcher struct {
	all     bool
	methods map[string]struct{}
}

func (m methodMatcher) matches(method string) bool {
	if m.all {
		return true
	}
	_, ok := m.methods[method]
	return ok
}

type bucket struct {
	mu           sync.Mutex
	name         string
	limit        rate.Limit
	burst        int
	limiter      *rate.Limiter
	matcher      methodMatcher
	blockedUntil time.Time // zero = not blocked; set only when the node told us its exact reset instant
}

func newBucket(name string, requestsPerMinute int) *bucket {
	limit, burst := rate.Inf, 0
	if requestsPerMinute > 0 {
		limit, burst = rate.Limit(float64(requestsPerMinute)/60.0), requestsPerMinute
	}
	return &bucket{
		name:    name,
		limit:   limit,
		burst:   burst,
		limiter: rate.NewLimiter(limit, burst),
	}
}

func (b *bucket) allow(now time.Time) (bool, time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if now.Before(b.blockedUntil) {
		return false, b.blockedUntil.Sub(now)
	}
	if !b.blockedUntil.IsZero() {
		// A known reset instant just passed: start clean, matching the node's own window
		b.limiter = rate.NewLimiter(b.limit, b.burst)
		b.blockedUntil = time.Time{}
	}

	if b.limiter.AllowN(now, 1) {
		return true, 0
	}

	reservation := b.limiter.ReserveN(now, 1)
	if !reservation.OK() {
		// Unreachable with the invariants of newBucket (a finite limit always has
		// burst >= 1), kept so a future change can't silently return a zero wait
		return false, b.refillInterval()
	}
	delay := reservation.DelayFrom(now)
	reservation.CancelAt(now) // only asking when a token frees up, not taking one
	return false, delay
}

func (b *bucket) notifyExceeded(now time.Time, resetAt *time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if resetAt != nil {
		b.blockedUntil = *resetAt
		return
	}
	b.limiter.ReserveN(now, b.burst) // deliberately not cancelled: forces a real deficit
}

// refillInterval is how long a single token takes to refill at the bucket's rate.
func (b *bucket) refillInterval() time.Duration {
	if b.limit == rate.Inf || b.limit <= 0 {
		return 0
	}
	return time.Duration(float64(time.Second) / float64(b.limit))
}
