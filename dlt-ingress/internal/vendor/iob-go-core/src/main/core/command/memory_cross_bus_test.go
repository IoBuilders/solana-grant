package command

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/observability/metrics"
)

type testCrossCmd struct{}

func (c *testCrossCmd) CrossCommandName() string { return "TestCrossCmd" }

type testCrossResp struct {
	event.BaseEvent
}

type testCrossErr struct {
	err error
}

func (e testCrossErr) Error() string { return e.err.Error() }
func (e testCrossErr) Unwrap() error { return e.err }

type okPort struct{}

func (p *okPort) Execute(ctx context.Context, cmd *testCrossCmd) (CrossResponse, error) {
	return &testCrossResp{BaseEvent: *event.NewBaseEvent()}, nil
}

type portReturnsError struct{}

func (p *portReturnsError) Execute(ctx context.Context, cmd *testCrossCmd) (CrossResponse, error) {
	return nil, testCrossErr{err: errors.New("boom")}
}

// portReturnsNonError exercises the "second return is not an error" branch of
// Dispatch. The CrossPort interface forbids this signature, so we inject the
// port directly into bus.ports to bypass RegisterPort.
type portReturnsNonError struct{}

func (p *portReturnsNonError) Execute(ctx context.Context, cmd *testCrossCmd) (CrossResponse, int) {
	return nil, 42
}

type portReturnsThreeValues struct{}

func (p *portReturnsThreeValues) Execute(ctx context.Context, cmd *testCrossCmd) (CrossResponse, error, int) {
	return &testCrossResp{BaseEvent: *event.NewBaseEvent()}, nil, 1
}

type invalidCrossCmd struct{}

type portWithInvalidCmdParam struct{}

func (p *portWithInvalidCmdParam) Execute(ctx context.Context, cmd *invalidCrossCmd) (CrossResponse, error) {
	return &testCrossResp{BaseEvent: *event.NewBaseEvent()}, nil
}

type portWrongParamCount struct{}

func (p *portWrongParamCount) Execute(ctx context.Context) (CrossResponse, error) {
	return &testCrossResp{BaseEvent: *event.NewBaseEvent()}, nil
}

type portWithoutExecuteMethod struct {
	X int
}

func TestCoreCrossCommandBusInMemory_RegisterPort_Errors(t *testing.T) {
	bus := NewCrossCommandBus(nil, metrics.NewRegistry())

	t.Run("port must be pointer to struct", func(t *testing.T) {
		err := bus.RegisterPort(okPort{})
		assert.Error(t, err)
	})

	t.Run("port must implement Execute", func(t *testing.T) {
		err := bus.RegisterPort(&portWithoutExecuteMethod{})
		assert.Error(t, err)
	})

	t.Run("execute method must have exactly (ctx, command)", func(t *testing.T) {
		err := bus.RegisterPort(&portWrongParamCount{})
		assert.Error(t, err)
	})

	t.Run("command parameter must implement CrossCommand", func(t *testing.T) {
		err := bus.RegisterPort(&portWithInvalidCmdParam{})
		assert.Error(t, err)
	})
}

func TestCoreCrossCommandBusInMemory_Dispatch_NoPort(t *testing.T) {
	bus := NewCrossCommandBus(nil, metrics.NewRegistry())

	_, err := bus.Dispatch(context.Background(), &testCrossCmd{})
	assert.NotNil(t, err)
}

func TestCoreCrossCommandBusInMemory_Dispatch_PortHasNoExecuteMethod(t *testing.T) {
	bus := NewCrossCommandBus(nil, metrics.NewRegistry())

	// Force an invalid port into the map (bypass RegisterPort)
	bus.ports["TestCrossCmd"] = &portWithoutExecuteMethod{}

	_, err := bus.Dispatch(context.Background(), &testCrossCmd{})
	assert.NotNil(t, err)
}

func TestCoreCrossCommandBusInMemory_Dispatch_UnexpectedResultCount(t *testing.T) {
	bus := NewCrossCommandBus(nil, metrics.NewRegistry())

	assert.NoError(t, bus.RegisterPort(&portReturnsThreeValues{}))

	_, err := bus.Dispatch(context.Background(), &testCrossCmd{})
	assert.NotNil(t, err)
}

func TestCoreCrossCommandBusInMemory_Dispatch_ReturnsError(t *testing.T) {
	bus := NewCrossCommandBus(nil, metrics.NewRegistry())

	assert.NoError(t, bus.RegisterPort(&portReturnsError{}))

	_, err := bus.Dispatch(context.Background(), &testCrossCmd{})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "boom")
}

func TestCoreCrossCommandBusInMemory_Dispatch_WhenSecondReturnIsNotError(t *testing.T) {
	bus := NewCrossCommandBus(nil, metrics.NewRegistry())
	bus.ports["TestCrossCmd"] = &portReturnsNonError{}

	_, err := bus.Dispatch(context.Background(), &testCrossCmd{})
	assert.Nil(t, err, "non-nilable second return must be silently ignored, not panic")
}
