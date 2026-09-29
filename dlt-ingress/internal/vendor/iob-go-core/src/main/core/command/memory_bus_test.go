package command

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/db"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/observability/metrics"
	"go.opentelemetry.io/otel/metric/noop"
)

type fakeTx struct {
	commitErr   error
	rollbackErr error

	commitCalls   int
	rollbackCalls int
}

func (t *fakeTx) Commit() error {
	t.commitCalls++
	return t.commitErr
}

func (t *fakeTx) Rollback() error {
	t.rollbackCalls++
	return t.rollbackErr
}

type fakeTM struct {
	tx db.Transaction
}

func (m *fakeTM) WithTransactionValue(parent context.Context) context.Context    { return parent }
func (m *fakeTM) WithNewTransactionValue(parent context.Context) context.Context { return parent }
func (m *fakeTM) TransactionValue(ctx context.Context) db.Transaction            { return m.tx }
func (m *fakeTM) HasTransactionValue(ctx context.Context) bool                   { return true }

type testCmd struct{}

func (c *testCmd) CommandName() string { return "TestCmd" }

type testResp struct {
	event.BaseEvent
}

// legacyHandlerErr satisfies the deprecated HandlerError interface, used to
// verify that bounded contexts still returning HandlerError continue to work.
type legacyHandlerErr struct {
	err error
}

func (e legacyHandlerErr) Error() string { return e.err.Error() }
func (e legacyHandlerErr) Unwrap() error { return e.err }

type okHandler struct{}

func (h *okHandler) Handle(ctx context.Context, cmd *testCmd) (Response, error) {
	return &testResp{BaseEvent: *event.NewBaseEvent()}, nil
}

type handlerReturnsError struct{}

func (h *handlerReturnsError) Handle(ctx context.Context, cmd *testCmd) (Response, error) {
	return &testResp{BaseEvent: *event.NewBaseEvent()}, errors.New("boom")
}

// handlerReturnsLegacyHandlerError exercises the permissive bus: handler still
// returns the deprecated HandlerError type but the bus must accept and surface
// it as a plain error.
type handlerReturnsLegacyHandlerError struct{}

func (h *handlerReturnsLegacyHandlerError) Handle(ctx context.Context, cmd *testCmd) (Response, error) {
	return &testResp{BaseEvent: *event.NewBaseEvent()}, legacyHandlerErr{err: errors.New("legacy-boom")}
}

type handlerReturnsThreeValues struct{}

func (h *handlerReturnsThreeValues) Handle(ctx context.Context, cmd *testCmd) (Response, error, int) {
	return &testResp{BaseEvent: *event.NewBaseEvent()}, nil, 1
}

type invalidCmd struct{}

type handlerWithInvalidCmdParam struct{}

func (h *handlerWithInvalidCmdParam) Handle(ctx context.Context, cmd *invalidCmd) (Response, error) {
	return &testResp{BaseEvent: *event.NewBaseEvent()}, nil
}

type handlerWrongParamCount struct{}

func (h *handlerWrongParamCount) Handle(ctx context.Context) (Response, error) {
	return &testResp{BaseEvent: *event.NewBaseEvent()}, nil
}

type handlerWithoutHandleMethod struct {
	X int
}

func TestMain(m *testing.M) {
	provider := noop.NewMeterProvider()
	noopMeter := provider.Meter("testMeter")
	metrics.SetDefaultMeter(noopMeter)
	os.Exit(m.Run())
}

func TestCoreCommandBusInMemory_RegisterHandler_Errors(t *testing.T) {
	bus := NewCommandBus(&fakeTM{tx: &fakeTx{}}, nil, metrics.NewRegistry())

	t.Run("handler must be pointer to struct", func(t *testing.T) {
		assert.Error(t, bus.RegisterHandler(okHandler{}))
	})

	t.Run("handler must implement Handle", func(t *testing.T) {
		assert.Error(t, bus.RegisterHandler(&handlerWithoutHandleMethod{}))
	})

	t.Run("handle method must have exactly (ctx, command)", func(t *testing.T) {
		assert.Error(t, bus.RegisterHandler(&handlerWrongParamCount{}))
	})

	t.Run("command parameter must implement Command interface", func(t *testing.T) {
		assert.Error(t, bus.RegisterHandler(&handlerWithInvalidCmdParam{}))
	})
}

func TestCoreCommandBusInMemory_Dispatch_NoHandler(t *testing.T) {
	bus := NewCommandBus(&fakeTM{tx: &fakeTx{}}, nil, metrics.NewRegistry())

	_, err := bus.Dispatch(context.Background(), &testCmd{})
	assert.Error(t, err)
}

func TestCoreCommandBusInMemory_Dispatch_HandlerHasNoHandleMethod(t *testing.T) {
	tx := &fakeTx{}
	bus := NewCommandBus(&fakeTM{tx: tx}, nil, metrics.NewRegistry())
	bus.handlers["TestCmd"] = &handlerWithoutHandleMethod{}

	_, err := bus.Dispatch(context.Background(), &testCmd{})
	assert.Error(t, err)
}

func TestCoreCommandBusInMemory_Dispatch_RollbackWhenUnexpectedResultCount(t *testing.T) {
	tx := &fakeTx{}
	bus := NewCommandBus(&fakeTM{tx: tx}, nil, metrics.NewRegistry())

	assert.NoError(t, bus.RegisterHandler(&handlerReturnsThreeValues{}))

	_, err := bus.Dispatch(context.Background(), &testCmd{})
	assert.Error(t, err)
	assert.Equal(t, 1, tx.rollbackCalls)
}

func TestCoreCommandBusInMemory_Dispatch_RollbackAndReturnError(t *testing.T) {
	tx := &fakeTx{}
	bus := NewCommandBus(&fakeTM{tx: tx}, nil, metrics.NewRegistry())

	assert.NoError(t, bus.RegisterHandler(&handlerReturnsError{}))

	_, err := bus.Dispatch(context.Background(), &testCmd{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
	assert.Equal(t, 1, tx.rollbackCalls)
}

func TestCoreCommandBusInMemory_Dispatch_AcceptsLegacyHandlerErrorReturn(t *testing.T) {
	tx := &fakeTx{}
	bus := NewCommandBus(&fakeTM{tx: tx}, nil, metrics.NewRegistry())

	assert.NoError(t, bus.RegisterHandler(&handlerReturnsLegacyHandlerError{}))

	_, err := bus.Dispatch(context.Background(), &testCmd{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "legacy-boom")
	assert.Equal(t, 1, tx.rollbackCalls)
}

func TestCoreCommandBusInMemory_Dispatch_CommitError(t *testing.T) {
	tx := &fakeTx{commitErr: errors.New("commit-failed")}
	bus := NewCommandBus(&fakeTM{tx: tx}, nil, metrics.NewRegistry())

	assert.NoError(t, bus.RegisterHandler(&okHandler{}))

	_, err := bus.Dispatch(context.Background(), &testCmd{})
	assert.Error(t, err)
	assert.Equal(t, 1, tx.commitCalls)
}
