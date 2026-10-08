package ratelimit

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

// fakeClock makes the token bucket's timing deterministic: tests advance time
// explicitly instead of sleeping.
type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func newFakeClock() *fakeClock {
	return &fakeClock{now: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)}
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

// allowedCalls counts how many consecutive calls the limiter lets through right
// now, up to max, without advancing the clock.
func allowedCalls(t *testing.T, limiter *RateLimiter, method string, max int) int {
	t.Helper()
	count := 0
	for i := 0; i < max; i++ {
		allowed, _ := limiter.Allow(context.Background(), method)
		if !allowed {
			break
		}
		count++
	}
	return count
}

func TestRateLimiter_Burst_Fires_A_Full_Minute_Budget_Immediately(t *testing.T) {
	clock := newFakeClock()
	limiter := New(withClock(clock.Now), WithBucket("TIER_1", 60, Methods("eth_call")))

	// burst = requestsPerMinute, so a whole minute's budget can leave at once
	assert.Equal(t, 60, allowedCalls(t, limiter, "eth_call", 100))

	allowed, retryAfter := limiter.Allow(context.Background(), "eth_call")

	assert.False(t, allowed)
	assert.Equal(t, 1*time.Second, retryAfter) // 60 req/min refills one token per second
}

func TestRateLimiter_Refills_Gradually_After_The_Burst(t *testing.T) {
	clock := newFakeClock()
	limiter := New(withClock(clock.Now), WithBucket("TIER_1", 60, Methods("eth_call")))
	assert.Equal(t, 60, allowedCalls(t, limiter, "eth_call", 60))

	clock.advance(1 * time.Second)
	assert.Equal(t, 1, allowedCalls(t, limiter, "eth_call", 10)) // one token refilled, not a full bucket

	clock.advance(10 * time.Second)
	assert.Equal(t, 10, allowedCalls(t, limiter, "eth_call", 60))

	clock.advance(1 * time.Hour)
	assert.Equal(t, 60, allowedCalls(t, limiter, "eth_call", 100)) // capped at the burst, no matter how long it idles
}

func TestRateLimiter_RetryAfter_Grows_With_The_Requested_Backlog(t *testing.T) {
	clock := newFakeClock()
	limiter := New(withClock(clock.Now), WithBucket("TIER_1", 60, Methods("eth_call")))
	assert.Equal(t, 60, allowedCalls(t, limiter, "eth_call", 60))

	// A rejected call must not consume a token, so every rejection reports the same wait
	for i := 0; i < 5; i++ {
		allowed, retryAfter := limiter.Allow(context.Background(), "eth_call")
		assert.False(t, allowed)
		assert.Equal(t, 1*time.Second, retryAfter)
	}
}

func TestRateLimiter_Unlimited_When_RequestsPerMinute_Not_Positive(t *testing.T) {
	// rate.Limit(0) would mean "never": the placeholder for a node whose real
	// quota is still unknown has to fail open instead
	for _, requestsPerMinute := range []int{0, -1} {
		clock := newFakeClock()
		limiter := New(withClock(clock.Now), WithBucket("DEFAULT", requestsPerMinute, MatchAll()))

		assert.Equal(t, rate.Inf, limiter.buckets[0].limit)
		assert.Equal(t, 1000, allowedCalls(t, limiter, "sendTransaction", 1000))

		allowed, retryAfter := limiter.Allow(context.Background(), "sendTransaction")
		assert.True(t, allowed)
		assert.Equal(t, time.Duration(0), retryAfter)
	}
}

func TestRateLimiter_Method_Matching_By_Bucket(t *testing.T) {
	clock := newFakeClock()
	limiter := New(
		withClock(clock.Now),
		WithBucket("TIER_1", 1, Methods("eth_call", "eth_sendRawTransaction")),
		WithBucket("TIER_3", 5, Methods("net_version")),
	)

	assert.Equal(t, 1, allowedCalls(t, limiter, "eth_call", 10))

	// Exhausting TIER_1 with eth_call also exhausts it for eth_sendRawTransaction
	allowed, _ := limiter.Allow(context.Background(), "eth_sendRawTransaction")
	assert.False(t, allowed)

	// ...but leaves other buckets untouched
	assert.Equal(t, 5, allowedCalls(t, limiter, "net_version", 10))
}

func TestRateLimiter_Unmatched_Method_Is_Always_Allowed(t *testing.T) {
	clock := newFakeClock()
	limiter := New(withClock(clock.Now), WithBucket("TIER_1", 1, Methods("eth_call")))
	assert.Equal(t, 1, allowedCalls(t, limiter, "eth_call", 10))

	// An unclassified method fails open rather than closed
	assert.Equal(t, 100, allowedCalls(t, limiter, "eth_getBalance", 100))
}

func TestRateLimiter_First_Matching_Bucket_Wins(t *testing.T) {
	clock := newFakeClock()
	limiter := New(
		withClock(clock.Now),
		WithBucket("TIER_1", 1, Methods("eth_call")),
		WithBucket("DEFAULT", 100, MatchAll()),
	)

	assert.Equal(t, 1, allowedCalls(t, limiter, "eth_call", 10)) // drawn from TIER_1, not the catch-all
	assert.Equal(t, 100, allowedCalls(t, limiter, "eth_getBalance", 200))
}

func TestRateLimiter_NotifyExceeded_With_Known_Reset_Blocks_Until_That_Instant(t *testing.T) {
	clock := newFakeClock()
	limiter := New(withClock(clock.Now), WithBucket("TIER_1", 60, Methods("eth_call")))
	resetAt := clock.Now().Add(30 * time.Second)

	// The node contradicts our own accounting: we thought there was capacity
	allowed, _ := limiter.Allow(context.Background(), "eth_call")
	assert.True(t, allowed)
	limiter.NotifyExceeded(context.Background(), "eth_call", &resetAt)

	allowed, retryAfter := limiter.Allow(context.Background(), "eth_call")
	assert.False(t, allowed)
	assert.Equal(t, 30*time.Second, retryAfter)

	clock.advance(10 * time.Second)
	allowed, retryAfter = limiter.Allow(context.Background(), "eth_call")
	assert.False(t, allowed)
	assert.Equal(t, 20*time.Second, retryAfter) // counts down towards the node's own reset

	// At the declared reset the bucket comes back full, matching the node's window
	clock.advance(20 * time.Second)
	assert.Equal(t, 60, allowedCalls(t, limiter, "eth_call", 100))
}

func TestRateLimiter_NotifyExceeded_With_Past_Reset_Recovers_Immediately(t *testing.T) {
	clock := newFakeClock()
	limiter := New(withClock(clock.Now), WithBucket("TIER_1", 60, Methods("eth_call")))
	assert.Equal(t, 60, allowedCalls(t, limiter, "eth_call", 60))
	resetAt := clock.Now().Add(-1 * time.Second)

	limiter.NotifyExceeded(context.Background(), "eth_call", &resetAt)

	// The node's declared window is already over, so its budget is back
	assert.Equal(t, 60, allowedCalls(t, limiter, "eth_call", 100))
}

func TestRateLimiter_NotifyExceeded_Without_Reset_Recovers_Gradually_Not_Fully(t *testing.T) {
	clock := newFakeClock()
	limiter := New(withClock(clock.Now), WithBucket("TIER_1", 60, Methods("eth_call")))

	// A live 429 with no usable header, on a bucket we believed was untouched
	limiter.NotifyExceeded(context.Background(), "eth_call", nil)

	// Regression guard: an earlier draft of this design reset the bucket to full
	// here, so the very next call went straight through
	allowed, retryAfter := limiter.Allow(context.Background(), "eth_call")
	assert.False(t, allowed)
	assert.Equal(t, 1*time.Second, retryAfter)

	clock.advance(1 * time.Second)
	assert.Equal(t, 1, allowedCalls(t, limiter, "eth_call", 60)) // one token, at the normal configured rate

	clock.advance(5 * time.Second)
	assert.Equal(t, 5, allowedCalls(t, limiter, "eth_call", 60))
}

func TestRateLimiter_NotifyExceeded_Without_Reset_Accounts_For_Already_Spent_Tokens(t *testing.T) {
	clock := newFakeClock()
	limiter := New(withClock(clock.Now), WithBucket("TIER_1", 60, Methods("eth_call")))
	assert.Equal(t, 30, allowedCalls(t, limiter, "eth_call", 30)) // half the burst already spent

	limiter.NotifyExceeded(context.Background(), "eth_call", nil)

	// Spending a full burst on a half-empty bucket leaves a real deficit, so
	// recovery starts later than a single token's refill
	allowed, retryAfter := limiter.Allow(context.Background(), "eth_call")
	assert.False(t, allowed)
	assert.Equal(t, 31*time.Second, retryAfter)

	clock.advance(30 * time.Second)
	allowed, _ = limiter.Allow(context.Background(), "eth_call")
	assert.False(t, allowed)

	clock.advance(1 * time.Second)
	assert.Equal(t, 1, allowedCalls(t, limiter, "eth_call", 60))
}

func TestRateLimiter_NotifyExceeded_On_Unlimited_Bucket_Without_Reset_Is_A_No_Op(t *testing.T) {
	clock := newFakeClock()
	limiter := New(withClock(clock.Now), WithBucket("DEFAULT", 0, MatchAll()))

	limiter.NotifyExceeded(context.Background(), "sendTransaction", nil)

	// Unlimited means no local accounting at all, so there is nothing to spend
	assert.Equal(t, 100, allowedCalls(t, limiter, "sendTransaction", 100))
}

func TestRateLimiter_NotifyExceeded_On_Unlimited_Bucket_Honours_Known_Reset(t *testing.T) {
	clock := newFakeClock()
	limiter := New(withClock(clock.Now), WithBucket("DEFAULT", 0, MatchAll()))
	resetAt := clock.Now().Add(10 * time.Second)

	limiter.NotifyExceeded(context.Background(), "sendTransaction", &resetAt)

	allowed, retryAfter := limiter.Allow(context.Background(), "sendTransaction")
	assert.False(t, allowed)
	assert.Equal(t, 10*time.Second, retryAfter)

	clock.advance(10 * time.Second)
	assert.Equal(t, 100, allowedCalls(t, limiter, "sendTransaction", 100))
}

func TestRateLimiter_NotifyExceeded_Unmatched_Method_Does_Nothing(t *testing.T) {
	clock := newFakeClock()
	limiter := New(withClock(clock.Now), WithBucket("TIER_1", 60, Methods("eth_call")))
	resetAt := clock.Now().Add(10 * time.Second)

	limiter.NotifyExceeded(context.Background(), "eth_getBalance", &resetAt)

	assert.Equal(t, 60, allowedCalls(t, limiter, "eth_call", 100))
}

func TestRateLimiter_Without_Buckets_Allows_Everything(t *testing.T) {
	limiter := New()

	allowed, retryAfter := limiter.Allow(context.Background(), "eth_call")

	assert.True(t, allowed)
	assert.Equal(t, time.Duration(0), retryAfter)
}

func TestRateLimiter_Concurrent_Callers_Share_One_Budget(t *testing.T) {
	limiter := New(WithBucket("TIER_1", 100, Methods("eth_call")))
	var mu sync.Mutex
	var wg sync.WaitGroup
	allowedCount := 0

	for i := 0; i < 500; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if allowed, _ := limiter.Allow(context.Background(), "eth_call"); allowed {
				mu.Lock()
				allowedCount++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	// The budget belongs to the node, so concurrency can't hand out more than one burst
	assert.LessOrEqual(t, allowedCount, 100)
	assert.GreaterOrEqual(t, allowedCount, 99) // one token may refill mid-run
}
