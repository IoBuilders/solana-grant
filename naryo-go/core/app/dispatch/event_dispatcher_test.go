//go:build test

package dispatch

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/logging"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

// --- stubs ---

type stubEvent struct{}

func (stubEvent) EventType() event.Type { return event.TypeBlock }
func (stubEvent) NodeID() uuid.UUID     { return uuid.Nil }

// recordingLogger counts Error calls; it is safe for concurrent use because
// triggers are processed in parallel goroutines.
type recordingLogger struct {
	mu     sync.Mutex
	errors int
}

func (l *recordingLogger) Debug(string, ...any)                         {}
func (l *recordingLogger) Info(string, ...any)                          {}
func (l *recordingLogger) Warn(string, ...any)                          {}
func (l *recordingLogger) Error(string, ...any)                         {}
func (l *recordingLogger) DebugWithCtx(context.Context, string, ...any) {}
func (l *recordingLogger) InfoWithCtx(context.Context, string, ...any)  {}
func (l *recordingLogger) WarnWithCtx(context.Context, string, ...any)  {}
func (l *recordingLogger) ErrorWithCtx(context.Context, string, ...any) {
	l.mu.Lock()
	l.errors++
	l.mu.Unlock()
}

func (l *recordingLogger) errorCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.errors
}

// fakeTrigger is a permanent trigger whose behavior is configured per test.
type fakeTrigger struct {
	supports  bool
	err       error
	panics    bool
	processed atomic.Int32
}

func (f *fakeTrigger) Supports(event.Event) bool { return f.supports }

func (f *fakeTrigger) Process(context.Context, event.Event) error {
	f.processed.Add(1)
	if f.panics {
		panic("boom")
	}
	return f.err
}

// fakeDisposableTrigger adds IsExpired, so it satisfies DisposableTrigger.
type fakeDisposableTrigger struct {
	fakeTrigger
	expired bool
}

func (f *fakeDisposableTrigger) IsExpired() bool { return f.expired }

// --- tests ---

func TestEventDispatcher_AddTrigger_SortsByType(t *testing.T) {
	d := NewEventDispatcher()

	d.AddTrigger(&fakeTrigger{})
	d.AddTrigger(&fakeDisposableTrigger{})

	assert.Len(t, d.permanentTriggers, 1)
	assert.Len(t, d.disposableTriggers, 1)
}

func TestEventDispatcher_Dispatch_InvokesAllSupportingTriggers(t *testing.T) {
	d := NewEventDispatcher()
	perm := &fakeTrigger{supports: true}
	disp := &fakeDisposableTrigger{fakeTrigger: fakeTrigger{supports: true}}
	d.AddTrigger(perm)
	d.AddTrigger(disp)

	d.Dispatch(context.Background(), stubEvent{})

	assert.Equal(t, int32(1), perm.processed.Load())
	assert.Equal(t, int32(1), disp.processed.Load())
}

func TestEventDispatcher_Dispatch_SkipsNonSupportingTriggers(t *testing.T) {
	d := NewEventDispatcher()
	trg := &fakeTrigger{supports: false}
	d.AddTrigger(trg)

	d.Dispatch(context.Background(), stubEvent{})

	assert.Equal(t, int32(0), trg.processed.Load())
}

// AC #1: a panicking trigger must not prevent the others from executing.
func TestEventDispatcher_Dispatch_PanicDoesNotStopSiblings(t *testing.T) {
	logger := &recordingLogger{}
	logging.SetDefaultLogger(logger)
	d := NewEventDispatcher()
	panicky := &fakeTrigger{supports: true, panics: true}
	healthy := &fakeTrigger{supports: true}
	d.AddTrigger(panicky)
	d.AddTrigger(healthy)

	assert.NotPanics(t, func() {
		d.Dispatch(context.Background(), stubEvent{})
	})

	assert.Equal(t, int32(1), healthy.processed.Load(), "sibling must still run")
	assert.GreaterOrEqual(t, logger.errorCount(), 1, "panic must be logged")
}

// A returned error is logged and swallowed; Dispatch returns normally.
func TestEventDispatcher_Dispatch_ProcessErrorIsLoggedNotPropagated(t *testing.T) {
	logger := &recordingLogger{}
	logging.SetDefaultLogger(logger)
	d := NewEventDispatcher()
	trg := &fakeTrigger{supports: true, err: errors.New("save failed")}
	d.AddTrigger(trg)

	assert.NotPanics(t, func() {
		d.Dispatch(context.Background(), stubEvent{})
	})

	assert.Equal(t, int32(1), trg.processed.Load())
	assert.Equal(t, 1, logger.errorCount())
}

// AC #2: expired disposable triggers are removed after Dispatch returns.
func TestEventDispatcher_Dispatch_RemovesExpiredDisposable(t *testing.T) {
	d := NewEventDispatcher()
	expired := &fakeDisposableTrigger{fakeTrigger: fakeTrigger{supports: true}, expired: true}
	alive := &fakeDisposableTrigger{fakeTrigger: fakeTrigger{supports: true}, expired: false}
	d.AddTrigger(expired)
	d.AddTrigger(alive)

	d.Dispatch(context.Background(), stubEvent{})

	require.Len(t, d.disposableTriggers, 1, "expired disposable must be dropped")
	survivor, ok := d.disposableTriggers[0].(*fakeDisposableTrigger)
	require.True(t, ok)
	assert.False(t, survivor.expired, "the surviving trigger must be the non-expired one")
}

func TestEventDispatcher_RemoveExpired_KeepsOnlyLiveTriggers(t *testing.T) {
	d := NewEventDispatcher()
	d.AddTrigger(&fakeDisposableTrigger{expired: true})
	d.AddTrigger(&fakeDisposableTrigger{expired: false})
	d.AddTrigger(&fakeDisposableTrigger{expired: true})

	d.RemoveExpired()

	assert.Len(t, d.disposableTriggers, 1)
}

func TestEventDispatcher_Dispatch_NoTriggers(t *testing.T) {
	d := NewEventDispatcher()

	assert.NotPanics(t, func() {
		d.Dispatch(context.Background(), stubEvent{})
	})
}

// Exercises the concurrent fan-out; meaningful under `go test -race`.
func TestEventDispatcher_Dispatch_ConcurrentFanOut(t *testing.T) {
	d := NewEventDispatcher()
	const triggers = 50
	all := make([]*fakeTrigger, triggers)
	for i := range all {
		all[i] = &fakeTrigger{supports: true}
		d.AddTrigger(all[i])
	}

	d.Dispatch(context.Background(), stubEvent{})

	for i, trg := range all {
		assert.Equal(t, int32(1), trg.processed.Load(), "trigger %d must have run once", i)
	}
}
