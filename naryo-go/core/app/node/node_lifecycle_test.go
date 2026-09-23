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

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/dispatch"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/subscription"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/node"
)

// --- stubs ---

// fakeNodeRunner is a hand-written test double for NodeRunner that lets
// tests control whether Run/Stop succeed or fail and records call counts,
// without driving a real subscriber/dispatcher.
type fakeNodeRunner struct {
	mu        sync.Mutex
	n         *node.Node
	running   bool
	runCalls  int
	stopCalls int
	runErr    error
	stopErr   error
}

func newFakeNodeRunner(id uuid.UUID) *fakeNodeRunner {
	return &fakeNodeRunner{n: &node.Node{ID: id}}
}

func (r *fakeNodeRunner) Run(context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.runCalls++
	if r.runErr != nil {
		return r.runErr
	}
	r.running = true
	return nil
}

func (r *fakeNodeRunner) Stop(context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stopCalls++
	if r.stopErr != nil {
		return r.stopErr
	}
	r.running = false
	return nil
}

func (r *fakeNodeRunner) IsRunning() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.running
}

func (r *fakeNodeRunner) GetNode() *node.Node                    { return r.n }
func (r *fakeNodeRunner) GetSusbcriber() subscription.Subscriber { return nil }
func (r *fakeNodeRunner) GetDispatcher() dispatch.Dispatcher     { return nil }

func (r *fakeNodeRunner) runCallCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.runCalls
}

func (r *fakeNodeRunner) stopCallCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.stopCalls
}

// --- tests ---

func TestDefaultNodeLifecycle_NewDefaultNodeLifecycle(t *testing.T) {
	l := NewDefaultNodeLifecycle()
	assert.False(t, l.IsRunning(uuid.New()))
}

func TestDefaultNodeLifecycle_LaunchOne_StartsAndTracksRunner(t *testing.T) {
	id := uuid.New()
	runner := newFakeNodeRunner(id)
	l := NewDefaultNodeLifecycle()

	err := l.LaunchOne(context.Background(), runner)

	require.NoError(t, err)
	assert.Equal(t, 1, runner.runCallCount())
	assert.True(t, l.IsRunning(id))
}

func TestDefaultNodeLifecycle_LaunchOne_AlreadyRunning_ReturnsErrorWithoutRerunning(t *testing.T) {
	id := uuid.New()
	runner := newFakeNodeRunner(id)
	l := NewDefaultNodeLifecycle()
	require.NoError(t, l.LaunchOne(context.Background(), runner))

	err := l.LaunchOne(context.Background(), runner)

	assert.ErrorContains(t, err, id.String())
	assert.Equal(t, 1, runner.runCallCount())
}

func TestDefaultNodeLifecycle_LaunchOne_RunError_PropagatesAndDoesNotTrack(t *testing.T) {
	id := uuid.New()
	runErr := errors.New("boom")
	runner := newFakeNodeRunner(id)
	runner.runErr = runErr
	l := NewDefaultNodeLifecycle()

	err := l.LaunchOne(context.Background(), runner)

	assert.ErrorIs(t, err, runErr)
	assert.False(t, l.IsRunning(id))
}

func TestDefaultNodeLifecycle_Launch_StartsEveryRunner(t *testing.T) {
	l := NewDefaultNodeLifecycle()
	r1 := newFakeNodeRunner(uuid.New())
	r2 := newFakeNodeRunner(uuid.New())

	err := l.Launch(context.Background(), []NodeRunner{r1, r2})

	require.NoError(t, err)
	assert.True(t, l.IsRunning(r1.GetNode().ID))
	assert.True(t, l.IsRunning(r2.GetNode().ID))
}

func TestDefaultNodeLifecycle_Launch_StopsAtFirstError(t *testing.T) {
	l := NewDefaultNodeLifecycle()
	runErr := errors.New("boom")
	r1 := newFakeNodeRunner(uuid.New())
	r1.runErr = runErr
	r2 := newFakeNodeRunner(uuid.New())

	err := l.Launch(context.Background(), []NodeRunner{r1, r2})

	assert.ErrorIs(t, err, runErr)
	assert.Equal(t, 0, r2.runCallCount())
}

func TestDefaultNodeLifecycle_IsRunning_UnknownNode_ReturnsFalse(t *testing.T) {
	l := NewDefaultNodeLifecycle()
	assert.False(t, l.IsRunning(uuid.New()))
}

func TestDefaultNodeLifecycle_IsRunning_TrackedButNoLongerRunning_ReturnsFalse(t *testing.T) {
	id := uuid.New()
	runner := newFakeNodeRunner(id)
	l := NewDefaultNodeLifecycle()
	require.NoError(t, l.LaunchOne(context.Background(), runner))

	require.NoError(t, runner.Stop(context.Background()))

	assert.False(t, l.IsRunning(id))
}

func TestDefaultNodeLifecycle_Stop_RunningNode_StopsRunner(t *testing.T) {
	id := uuid.New()
	runner := newFakeNodeRunner(id)
	l := NewDefaultNodeLifecycle()
	require.NoError(t, l.LaunchOne(context.Background(), runner))

	err := l.Stop(context.Background(), id)

	require.NoError(t, err)
	assert.Equal(t, 1, runner.stopCallCount())
	assert.False(t, l.IsRunning(id))
}

func TestDefaultNodeLifecycle_Stop_NotRunning_ReturnsError(t *testing.T) {
	l := NewDefaultNodeLifecycle()
	id := uuid.New()

	err := l.Stop(context.Background(), id)

	assert.ErrorContains(t, err, id.String())
}

func TestDefaultNodeLifecycle_Stop_RunnerStopError_Propagates(t *testing.T) {
	id := uuid.New()
	stopErr := errors.New("boom")
	runner := newFakeNodeRunner(id)
	runner.stopErr = stopErr
	l := NewDefaultNodeLifecycle()
	require.NoError(t, l.LaunchOne(context.Background(), runner))

	err := l.Stop(context.Background(), id)

	assert.ErrorIs(t, err, stopErr)
}

func TestDefaultNodeLifecycle_Restart_StopsOldRunnerAndLaunchesNew(t *testing.T) {
	id := uuid.New()
	oldRunner := newFakeNodeRunner(id)
	newRunner := newFakeNodeRunner(id)
	l := NewDefaultNodeLifecycle()
	require.NoError(t, l.LaunchOne(context.Background(), oldRunner))

	err := l.Restart(context.Background(), id, newRunner)

	require.NoError(t, err)
	assert.Equal(t, 1, oldRunner.stopCallCount())
	assert.Equal(t, 1, newRunner.runCallCount())
	assert.True(t, l.IsRunning(id))
}

func TestDefaultNodeLifecycle_Restart_NotRunning_ReturnsErrorWithoutLaunching(t *testing.T) {
	id := uuid.New()
	newRunner := newFakeNodeRunner(id)
	l := NewDefaultNodeLifecycle()

	err := l.Restart(context.Background(), id, newRunner)

	assert.Error(t, err)
	assert.Equal(t, 0, newRunner.runCallCount())
}
