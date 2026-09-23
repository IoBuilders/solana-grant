//go:build test

package node

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/trigger"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/node"
)

// --- stubs ---

// fakeSubscriber is a hand-written test double for subscription.Subscriber
// that records how many times Subscribe was called and the context it was
// last called with, so tests can assert Run/Stop wiring without a real
// subscription loop.
type fakeSubscriber struct {
	mu    sync.Mutex
	calls int
	ctx   context.Context
}

func (s *fakeSubscriber) Subscribe(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	s.ctx = ctx
}

func (s *fakeSubscriber) subscribeCalls() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls
}

func (s *fakeSubscriber) subscribedCtx() context.Context {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ctx
}

// fakeDispatcher is a hand-written test double for dispatch.Dispatcher; it
// only needs to satisfy the interface here, its behavior is never exercised.
type fakeDispatcher struct{}

func (d *fakeDispatcher) Dispatch(context.Context, event.Event) {}
func (d *fakeDispatcher) AddTrigger(trigger.Trigger)            {}
func (d *fakeDispatcher) RemoveExpired()                        {}

// --- tests ---

func TestDefaultNodeRunner_NewDefaultNodeRunner(t *testing.T) {
	n := &node.Node{ID: uuid.New()}
	subscriber := &fakeSubscriber{}
	dispatcher := &fakeDispatcher{}

	r := NewDefaultNodeRunner(n, subscriber, dispatcher)

	assert.Same(t, n, r.GetNode())
	assert.Same(t, subscriber, r.GetSusbcriber())
	assert.Same(t, dispatcher, r.GetDispatcher())
	assert.False(t, r.IsRunning())
}

func TestDefaultNodeRunner_Run_SubscribesAndMarksRunning(t *testing.T) {
	subscriber := &fakeSubscriber{}
	r := NewDefaultNodeRunner(&node.Node{ID: uuid.New()}, subscriber, &fakeDispatcher{})

	err := r.Run(context.Background())

	require.NoError(t, err)
	assert.True(t, r.IsRunning())
	assert.Equal(t, 1, subscriber.subscribeCalls())
}

func TestDefaultNodeRunner_Run_AlreadyRunning_DoesNotResubscribe(t *testing.T) {
	subscriber := &fakeSubscriber{}
	r := NewDefaultNodeRunner(&node.Node{ID: uuid.New()}, subscriber, &fakeDispatcher{})
	require.NoError(t, r.Run(context.Background()))

	err := r.Run(context.Background())

	require.NoError(t, err)
	assert.True(t, r.IsRunning())
	assert.Equal(t, 1, subscriber.subscribeCalls())
}

func TestDefaultNodeRunner_Stop_CancelsSubscriberContextAndMarksNotRunning(t *testing.T) {
	subscriber := &fakeSubscriber{}
	r := NewDefaultNodeRunner(&node.Node{ID: uuid.New()}, subscriber, &fakeDispatcher{})
	require.NoError(t, r.Run(context.Background()))

	err := r.Stop(context.Background())

	require.NoError(t, err)
	assert.False(t, r.IsRunning())
	require.NotNil(t, subscriber.subscribedCtx())
	assert.ErrorIs(t, subscriber.subscribedCtx().Err(), context.Canceled)
}

func TestDefaultNodeRunner_Stop_NotRunning_IsNoop(t *testing.T) {
	subscriber := &fakeSubscriber{}
	r := NewDefaultNodeRunner(&node.Node{ID: uuid.New()}, subscriber, &fakeDispatcher{})

	err := r.Stop(context.Background())

	require.NoError(t, err)
	assert.False(t, r.IsRunning())
	assert.Equal(t, 0, subscriber.subscribeCalls())
}

func TestDefaultNodeRunner_RunStopRun_ResubscribesWithFreshContext(t *testing.T) {
	subscriber := &fakeSubscriber{}
	r := NewDefaultNodeRunner(&node.Node{ID: uuid.New()}, subscriber, &fakeDispatcher{})
	require.NoError(t, r.Run(context.Background()))
	require.NoError(t, r.Stop(context.Background()))

	err := r.Run(context.Background())

	require.NoError(t, err)
	assert.True(t, r.IsRunning())
	assert.Equal(t, 2, subscriber.subscribeCalls())
	assert.NoError(t, subscriber.subscribedCtx().Err())
}
