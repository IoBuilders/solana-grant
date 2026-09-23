//go:build test

package subscription

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/logging"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/retry"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/trigger"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/common"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/node"
)

// --- stubs ---

// fakeDispatcher is a hand-written test double for dispatch.Dispatcher.
// onDispatch, if set, is invoked synchronously after recording each event,
// letting tests cancel the subscriber's context from inside the loop it drives.
type fakeDispatcher struct {
	mu         sync.Mutex
	events     []event.Event
	onDispatch func(e event.Event)
}

func (d *fakeDispatcher) Dispatch(_ context.Context, e event.Event) {
	d.mu.Lock()
	d.events = append(d.events, e)
	cb := d.onDispatch
	d.mu.Unlock()
	if cb != nil {
		cb(e)
	}
}

func (d *fakeDispatcher) AddTrigger(trigger.Trigger) {}

func (d *fakeDispatcher) RemoveExpired() {}

func (d *fakeDispatcher) dispatched() []event.Event {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]event.Event(nil), d.events...)
}

func (d *fakeDispatcher) dispatchedSlots() []uint64 {
	events := d.dispatched()
	slots := make([]uint64, len(events))
	for i, e := range events {
		slots[i] = e.(event.SlotEvent).Slot
	}
	return slots
}

// fastRetryConfiguration makes CustomRetryer fail immediately (no sleep)
// after a single attempt: CustomRetryer requires delay/maxDelay of at least
// a second, but only sleeps between attempts, and MaxRetries: 1 means there
// is no attempt after the first to sleep before.
func fastRetryConfiguration() *common.RetryConfiguration {
	return &common.RetryConfiguration{
		MaxRetries:   1,
		InitialDelay: time.Second,
		MaxDelay:     time.Second,
		Multiplier:   1,
	}
}

// newPollNode builds a Node configured for POLL subscription with the given
// interval, which is all handleSlots' type assertion on MethodConfiguration needs.
func newPollNode(t *testing.T, interval time.Duration) *node.Node {
	t.Helper()
	method, err := node.NewPollBlockSubscriptionMethodConfiguration(interval)
	require.NoError(t, err)
	subscription, err := node.NewBlockSubscriptionConfiguration(method, 0)
	require.NoError(t, err)
	return &node.Node{ID: uuid.New(), Subscription: subscription}
}

// --- tests ---

func TestSolanaPollSlotSubscriber_HandleSlots_StartSlotZero_UsesStartSlotCalculator(t *testing.T) {
	store := &fakeLatestBlockStore{}
	calc := NewSolanaStartSlotCalculator(&node.Node{ID: uuid.New()}, &fakeBlockInteractor{slot: 500}, store)

	dispatcher := &fakeDispatcher{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dispatcher.onDispatch = func(event.Event) { cancel() }

	s := NewSolanaPollSlotSubscriber(newPollNode(t, time.Hour), &fakeBlockInteractor{}, dispatcher, calc, retry.NewCustomRetryer(), fastRetryConfiguration())

	s.handleSlots(ctx)

	assert.Equal(t, []uint64{500}, dispatcher.dispatchedSlots())
	assert.Equal(t, uint64(501), s.slot)
}

func TestSolanaPollSlotSubscriber_HandleSlots_StartSlotCalculatorError_StopsLoop(t *testing.T) {
	dbErr := errors.New("connection reset")
	store := &fakeLatestBlockStore{latestErr: dbErr}
	calc := NewSolanaStartSlotCalculator(&node.Node{ID: uuid.New()}, &fakeBlockInteractor{}, store)

	dispatcher := &fakeDispatcher{}
	s := NewSolanaPollSlotSubscriber(newPollNode(t, time.Hour), &fakeBlockInteractor{}, dispatcher, calc, retry.NewCustomRetryer(), fastRetryConfiguration())

	done := make(chan struct{})
	go func() {
		s.handleSlots(context.Background())
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("handleSlots did not return after a start-slot-calculator error")
	}

	assert.Empty(t, dispatcher.dispatched())
	assert.Equal(t, uint64(0), s.slot)
}

func TestSolanaPollSlotSubscriber_HandleSlots_DispatchesSequentialSlotsWhileCaughtUp(t *testing.T) {
	dispatcher := &fakeDispatcher{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dispatcher.onDispatch = func(event.Event) {
		if len(dispatcher.dispatched()) >= 3 {
			cancel()
		}
	}

	s := NewSolanaPollSlotSubscriber(newPollNode(t, time.Hour), &fakeBlockInteractor{slot: 15}, dispatcher, nil, retry.NewCustomRetryer(), fastRetryConfiguration())
	s.slot = 10

	s.handleSlots(ctx)

	assert.Equal(t, []uint64{10, 11, 12}, dispatcher.dispatchedSlots())
	assert.Equal(t, uint64(13), s.slot)
}

func TestSolanaPollSlotSubscriber_HandleSlots_SlotAheadOfChain_WaitsThenStopsOnContextCancel(t *testing.T) {
	dispatcher := &fakeDispatcher{}
	s := NewSolanaPollSlotSubscriber(newPollNode(t, time.Hour), &fakeBlockInteractor{slot: 50}, dispatcher, nil, retry.NewCustomRetryer(), fastRetryConfiguration())
	s.slot = 100

	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(50*time.Millisecond, cancel)

	done := make(chan struct{})
	go func() {
		s.handleSlots(ctx)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("handleSlots did not stop on context cancellation while waiting for the chain to catch up")
	}

	assert.Empty(t, dispatcher.dispatched())
	assert.Equal(t, uint64(100), s.slot)
}

func TestSolanaPollSlotSubscriber_HandleSlots_BlockInteractorError_StopsLoop(t *testing.T) {
	dispatcher := &fakeDispatcher{}
	interactorErr := errors.New("rpc failed")
	s := NewSolanaPollSlotSubscriber(newPollNode(t, time.Hour), &fakeBlockInteractor{slotErr: interactorErr}, dispatcher, nil, retry.NewCustomRetryer(), fastRetryConfiguration())
	s.slot = 5

	done := make(chan struct{})
	go func() {
		s.handleSlots(context.Background())
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("handleSlots did not return after a block interactor error")
	}

	assert.Empty(t, dispatcher.dispatched())
	assert.Equal(t, uint64(5), s.slot)
}

func TestSolanaPollSlotSubscriber_HandleSlots_ContextAlreadyCanceled_ReturnsImmediately(t *testing.T) {
	dispatcher := &fakeDispatcher{}
	s := NewSolanaPollSlotSubscriber(newPollNode(t, time.Hour), &fakeBlockInteractor{}, dispatcher, nil, retry.NewCustomRetryer(), fastRetryConfiguration())
	s.slot = 5

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	s.handleSlots(ctx)

	assert.Empty(t, dispatcher.dispatched())
	assert.Equal(t, uint64(5), s.slot)
}

func TestSolanaPollSlotSubscriber_Subscribe_RunsHandleSlotsAsynchronously(t *testing.T) {
	// A recording logger lets the test wait for the "Subscription stopped" log
	// that closes out Subscribe's goroutine, so it never outlives the test and
	// races the next test's use of the (unsynchronized) default logger global.
	logger := &recordingLogger{}
	logging.SetDefaultLogger(logger)
	t.Cleanup(func() { logging.SetDefaultLogger(logging.NewSlogLogger(slog.Default())) })

	dispatcher := &fakeDispatcher{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dispatcher.onDispatch = func(event.Event) { cancel() }

	s := NewSolanaPollSlotSubscriber(newPollNode(t, time.Hour), &fakeBlockInteractor{slot: 15}, dispatcher, nil, retry.NewCustomRetryer(), fastRetryConfiguration())
	s.slot = 10

	s.Subscribe(ctx)

	require.Eventually(t, func() bool { return logger.infoCount() >= 2 }, time.Second, 10*time.Millisecond,
		"handleSlots must log both its start and its stop before the goroutine ends")
	assert.NotEmpty(t, dispatcher.dispatchedSlots())
}

func TestSolanaPollSlotSubscriber_Subscribe_RecoversFromPanic(t *testing.T) {
	logger := &recordingLogger{}
	logging.SetDefaultLogger(logger)
	t.Cleanup(func() { logging.SetDefaultLogger(logging.NewSlogLogger(slog.Default())) })

	dispatcher := &fakeDispatcher{}
	// startSlotCalculator is left nil: with slot == 0, handleSlots dereferences
	// it to fetch the start slot, panicking. Subscribe must recover it.
	s := NewSolanaPollSlotSubscriber(newPollNode(t, time.Hour), &fakeBlockInteractor{}, dispatcher, nil, retry.NewCustomRetryer(), fastRetryConfiguration())

	assert.NotPanics(t, func() {
		s.Subscribe(context.Background())
		require.Eventually(t, func() bool { return logger.errorCount() >= 1 }, time.Second, 10*time.Millisecond,
			"panic recovery must be logged")
	})
}

// recordingLogger is a hand-written test double for logging.Logger that
// counts Info/InfoWithCtx and Error/ErrorWithCtx calls.
type recordingLogger struct {
	mu     sync.Mutex
	infos  int
	errors int
}

func (l *recordingLogger) Debug(string, ...any) {}
func (l *recordingLogger) Info(string, ...any) {
	l.mu.Lock()
	l.infos++
	l.mu.Unlock()
}
func (l *recordingLogger) Warn(string, ...any) {}
func (l *recordingLogger) Error(string, ...any) {
	l.mu.Lock()
	l.errors++
	l.mu.Unlock()
}
func (l *recordingLogger) DebugWithCtx(context.Context, string, ...any) {}
func (l *recordingLogger) InfoWithCtx(context.Context, string, ...any) {
	l.mu.Lock()
	l.infos++
	l.mu.Unlock()
}
func (l *recordingLogger) WarnWithCtx(context.Context, string, ...any) {}
func (l *recordingLogger) ErrorWithCtx(context.Context, string, ...any) {
	l.mu.Lock()
	l.errors++
	l.mu.Unlock()
}

func (l *recordingLogger) infoCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.infos
}

func (l *recordingLogger) errorCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.errors
}
