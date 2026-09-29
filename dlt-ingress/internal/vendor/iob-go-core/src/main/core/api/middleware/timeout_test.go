package middleware_test

// Naming convention:
//   TestCoreTimeoutMiddleware_*       behavior tests, run in `test-core` CI job
//   TestRaceCoreTimeoutMiddleware_*   race-specific tests, run in `race-core` CI job under -race
//
// Local run:
//   go test       -count=1 -run "^TestCoreTimeoutMiddleware"     ./src/main/core/api/middleware/...
//   go test -race -count=1 -run "^TestRaceCoreTimeoutMiddleware" ./src/main/core/api/middleware/...

import (
	"context"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api/middleware"
)

func newRouter(timeout time.Duration, handler gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.TimeOutMiddleware(timeout))
	r.GET("/probe", handler)
	return r
}

func doRequest(r *gin.Engine) (*httptest.ResponseRecorder, time.Duration) {
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	w := httptest.NewRecorder()
	start := time.Now()
	r.ServeHTTP(w, req)
	return w, time.Since(start)
}

func TestCoreTimeoutMiddleware_HappyPath(t *testing.T) {
	r := newRouter(200*time.Millisecond, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w, elapsed := doRequest(r)

	assert.Equal(t, http.StatusOK, w.Code, "fast handler must return 200")
	assert.Contains(t, w.Body.String(), `"ok":true`)
	assert.Less(t, elapsed, 50*time.Millisecond,
		"fast handler must not get close to the timeout")
}

// -----------------------------------------------------------------------------
// Bug Data race on *gin.Context.
//
// The current middleware passes the original `c` to a goroutine and, when the
// timeout fires, the outer goroutine calls AbortWithStatusJSON (which writes
// c.index) while the orphan goroutine continues inside c.Next() (which also
// mutates c.index). No Copy(), no lock: guaranteed race.
// -----------------------------------------------------------------------------

func TestRaceCoreTimeoutMiddleware_NoDataRaceOnGinContext(t *testing.T) {
	const (
		timeout     = 20 * time.Millisecond
		handlerWork = 80 * time.Millisecond
	)

	handlerDone := make(chan struct{})
	r := newRouter(timeout, func(c *gin.Context) {
		c.Set("phase", "before-sleep") // first write to c.Keys
		time.Sleep(handlerWork)        // ensures the timeout fires first
		c.Set("phase", "after-sleep")  // second write: concurrent with the outer goroutine
		close(handlerDone)
	})

	doRequest(r)

	// Give the orphan goroutine time to finish its late writes;
	// the race detector will already have recorded the event before this point.
	select {
	case <-handlerDone:
	case <-time.After(handlerWork + 200*time.Millisecond):
		t.Fatal("the orphan goroutine never finished: probable bug goroutine leak")
	}
}

// -----------------------------------------------------------------------------
// Bug Race on c.Writer.
//
// Gin's ResponseWriter is NOT thread-safe. With the current middleware, after
// the timeout, the outer goroutine writes the 504 while the orphan goroutine may
// be writing headers/body from the handler.
// -----------------------------------------------------------------------------

func TestRaceCoreTimeoutMiddleware_NoRaceOnResponseWriter(t *testing.T) {
	const (
		timeout     = 20 * time.Millisecond
		handlerWork = 80 * time.Millisecond
	)

	handlerDone := make(chan struct{})
	r := newRouter(timeout, func(c *gin.Context) {
		time.Sleep(handlerWork)
		// Late writes to the writer: concurrent with the outer goroutine's
		// AbortWithStatusJSON call in the buggy implementation.
		c.Header("X-Late-Header", "1")
		c.JSON(http.StatusOK, gin.H{"late": true})
		close(handlerDone)
	})

	doRequest(r)

	select {
	case <-handlerDone:
	case <-time.After(handlerWork + 200*time.Millisecond):
		t.Fatal("the orphan goroutine never finished: probable bug goroutine leak")
	}
}

// -----------------------------------------------------------------------------
// Bug Goroutine leak.
//
// The `finished` channel is created without a buffer and, on the timeout path,
// nobody keeps reading from it. When the orphan goroutine tries to send
// `finished <- struct{}{}` after finishing its work, it stays blocked forever.
// Each timed-out request = one leaked goroutine.
// -----------------------------------------------------------------------------

func TestCoreTimeoutMiddleware_NoGoroutineLeak(t *testing.T) {
	const (
		iterations  = 50
		timeout     = 10 * time.Millisecond
		handlerWork = 30 * time.Millisecond
	)

	// Handler that finishes its work (the orphan goroutine will try to write
	// to the unbuffered channel and will hang in the current implementation).
	r := newRouter(timeout, func(c *gin.Context) {
		time.Sleep(handlerWork)
	})

	// Stabilize the counter before measuring.
	runtime.GC()
	runtime.GC()
	time.Sleep(50 * time.Millisecond)
	before := runtime.NumGoroutine()

	for i := 0; i < iterations; i++ {
		doRequest(r)
	}

	// Wait for any healthy goroutine to finish before measuring.
	time.Sleep(handlerWork + 100*time.Millisecond)
	runtime.GC()
	runtime.GC()
	after := runtime.NumGoroutine()

	delta := after - before
	// Generous margin for runtime noise (GC, http.Server housekeeping).
	const maxAcceptableDelta = 5
	assert.LessOrEqual(t, delta, maxAcceptableDelta,
		"%d goroutines leaked after %d timed-out requests (expected <= %d)",
		delta, iterations, maxAcceptableDelta)
}

// -----------------------------------------------------------------------------
// Test Handler that respects ctx is canceled quickly.
//
// This is the GOOD path: when the handler checks ctx.Done(), cooperative
// cancellation works and we return 504 close to the timeout
// Regression guard.
// -----------------------------------------------------------------------------

func TestCoreTimeoutMiddleware_RespectfulHandlerCanceledFast(t *testing.T) {
	const (
		timeout      = 30 * time.Millisecond
		handlerBlock = 1 * time.Second
	)

	r := newRouter(timeout, func(c *gin.Context) {
		select {
		case <-time.After(handlerBlock): // should never be reached
		case <-c.Request.Context().Done():
			// the handler respects ctx: aborts cleanly
		}
	})

	w, elapsed := doRequest(r)

	assert.Equal(t, http.StatusGatewayTimeout, w.Code)
	assert.Less(t, elapsed, timeout+50*time.Millisecond,
		"a respectful handler must return 504 close to the timeout")
}

// -----------------------------------------------------------------------------
// Bug Handler that ignores ctx is NOT magically canceled.
//
// This test documents the actual cooperative nature of cancellation in Go.
//
// Current implementation: it lies. It returns 504 at around ~timeout ms even
// though the handler keeps running in an orphan goroutine for much longer.
// The test detects the lie: the elapsed time seen by the client is ~timeout,
// not ~handlerWork.
//
// -----------------------------------------------------------------------------

func TestCoreTimeoutMiddleware_DisrespectfulHandlerNotMagicallyCanceled(t *testing.T) {
	const (
		timeout     = 20 * time.Millisecond
		handlerWork = 200 * time.Millisecond
	)

	var handlerCompletedAt atomic.Int64
	r := newRouter(timeout, func(c *gin.Context) {
		time.Sleep(handlerWork) // intentionally ignores c.Request.Context()
		handlerCompletedAt.Store(time.Now().UnixNano())
	})

	w, elapsed := doRequest(r)

	require.Equal(t, http.StatusGatewayTimeout, w.Code,
		"timeout must respond 504 even with a disrespectful handler")

	// Key assertion: the client must observe the handler's REAL latency,
	// not a fake ~timeout latency. If elapsed < handlerWork/2, it means the
	// middleware is responding before the handler finishes — exactly the lie
	assert.GreaterOrEqual(t, elapsed, handlerWork-50*time.Millisecond,
		"middleware is lying about cancellation: it responded in %v "+
			"but the handler takes %v and does not respect ctx. The response should "+
			"have been delayed until the handler finished (honest cooperative cancellation).",
		elapsed, handlerWork)
}

// -----------------------------------------------------------------------------
// Test Timeout response contract: 504 with structured JSON body.
//
// Verifies that the client-visible behavior does not change
// Regression guard.
// -----------------------------------------------------------------------------

func TestCoreTimeoutMiddleware_ResponseShape_504WithJSONBody(t *testing.T) {
	r := newRouter(20*time.Millisecond, func(c *gin.Context) {
		select {
		case <-time.After(1 * time.Second):
		case <-c.Request.Context().Done():
		}
	})

	w, _ := doRequest(r)

	assert.Equal(t, http.StatusGatewayTimeout, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	body := w.Body.String()
	assert.True(t,
		strings.Contains(body, "timeout") || strings.Contains(body, "Timeout") ||
			strings.Contains(body, "TIMEOUT"),
		"body must mention timeout; got: %s", body)
}

// -----------------------------------------------------------------------------
// Test: The middleware must propagate a ctx with deadline to c.Request.Context().
//
// Regression guard.
// -----------------------------------------------------------------------------

func TestCoreTimeoutMiddleware_ContextDeadlinePropagation(t *testing.T) {
	const timeout = 100 * time.Millisecond

	var (
		hasDeadline   bool
		deadline      time.Time
		ctxErrAtStart error
	)
	r := newRouter(timeout, func(c *gin.Context) {
		dl, ok := c.Request.Context().Deadline()
		hasDeadline = ok
		deadline = dl
		ctxErrAtStart = c.Request.Context().Err()
		c.Status(http.StatusOK)
	})

	doRequest(r)

	assert.True(t, hasDeadline, "the handler must see a deadline in its ctx")
	assert.Nil(t, ctxErrAtStart, "the ctx must be alive when the handler starts")
	require.True(t, hasDeadline)
	assert.WithinDuration(t, time.Now().Add(timeout), deadline, 50*time.Millisecond,
		"the deadline must approximately match now+timeout")

	// Also verify that the ctx is canceled after the deadline.
	dummy, dummyCancel := context.WithDeadline(context.Background(), deadline)
	defer dummyCancel()
	<-dummy.Done() // blocks until the deadline, confirming it is real
	assert.Error(t, dummy.Err())
}
