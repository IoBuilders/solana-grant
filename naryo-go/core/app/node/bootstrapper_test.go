//go:build test

package node

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- stubs ---

// fakeNodeInitializer is a hand-written test double for NodeInitializer that
// returns a canned set of runners or a canned error, and records how many
// times Initialize was called.
type fakeNodeInitializer struct {
	mu      sync.Mutex
	calls   int
	runners []NodeRunner
	err     error
}

func (i *fakeNodeInitializer) Initialize(context.Context) ([]NodeRunner, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.calls++
	if i.err != nil {
		return nil, i.err
	}
	return i.runners, nil
}

func (i *fakeNodeInitializer) callCount() int {
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.calls
}

// fakeNodeLifecycle is a hand-written test double for NodeLifecycle; only
// Launch is exercised by DefaultBootstrapper, the rest satisfy the interface.
type fakeNodeLifecycle struct {
	mu              sync.Mutex
	launchCalls     int
	launchedRunners []NodeRunner
	launchErr       error
}

func (l *fakeNodeLifecycle) Launch(_ context.Context, nodeRunners []NodeRunner) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.launchCalls++
	l.launchedRunners = nodeRunners
	return l.launchErr
}

func (l *fakeNodeLifecycle) LaunchOne(context.Context, NodeRunner) error { return nil }
func (l *fakeNodeLifecycle) IsRunning(uuid.UUID) bool                    { return false }
func (l *fakeNodeLifecycle) Stop(context.Context, uuid.UUID) error       { return nil }
func (l *fakeNodeLifecycle) Restart(context.Context, uuid.UUID, NodeRunner) error {
	return nil
}

func (l *fakeNodeLifecycle) launchCallCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.launchCalls
}

// --- tests ---

func TestDefaultBootstrapper_NewDefaultBootstrapper(t *testing.T) {
	b := NewDefaultBootstrapper(&fakeNodeInitializer{}, &fakeNodeLifecycle{})
	assert.False(t, b.IsStarted(context.Background()))
}

func TestDefaultBootstrapper_Start_InitializesAndLaunchesNodes(t *testing.T) {
	runners := []NodeRunner{newFakeNodeRunner(uuid.New()), newFakeNodeRunner(uuid.New())}
	initializer := &fakeNodeInitializer{runners: runners}
	lifecycle := &fakeNodeLifecycle{}
	b := NewDefaultBootstrapper(initializer, lifecycle)

	err := b.Start(context.Background())

	require.NoError(t, err)
	assert.True(t, b.IsStarted(context.Background()))
	assert.Equal(t, 1, initializer.callCount())
	assert.Equal(t, 1, lifecycle.launchCallCount())
	assert.Equal(t, runners, lifecycle.launchedRunners)
}

func TestDefaultBootstrapper_Start_AlreadyStarted_IsNoop(t *testing.T) {
	initializer := &fakeNodeInitializer{}
	lifecycle := &fakeNodeLifecycle{}
	b := NewDefaultBootstrapper(initializer, lifecycle)
	require.NoError(t, b.Start(context.Background()))

	err := b.Start(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 1, initializer.callCount())
	assert.Equal(t, 1, lifecycle.launchCallCount())
}

func TestDefaultBootstrapper_Start_InitializeError_PropagatesAndDoesNotStart(t *testing.T) {
	initErr := errors.New("boom")
	initializer := &fakeNodeInitializer{err: initErr}
	lifecycle := &fakeNodeLifecycle{}
	b := NewDefaultBootstrapper(initializer, lifecycle)

	err := b.Start(context.Background())

	assert.ErrorIs(t, err, initErr)
	assert.False(t, b.IsStarted(context.Background()))
	assert.Equal(t, 0, lifecycle.launchCallCount())
}

func TestDefaultBootstrapper_Start_LaunchError_PropagatesAndDoesNotStart(t *testing.T) {
	launchErr := errors.New("boom")
	initializer := &fakeNodeInitializer{}
	lifecycle := &fakeNodeLifecycle{launchErr: launchErr}
	b := NewDefaultBootstrapper(initializer, lifecycle)

	err := b.Start(context.Background())

	assert.ErrorIs(t, err, launchErr)
	assert.False(t, b.IsStarted(context.Background()))
}
