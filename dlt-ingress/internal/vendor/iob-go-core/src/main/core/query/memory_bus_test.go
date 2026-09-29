package query

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testQuery struct {
	Value string
}

func (q *testQuery) QueryName() string {
	t := reflect.TypeOf(*q)
	return fmt.Sprintf("%s.%s", t.PkgPath(), t.Name())
}

type okHandler struct{}

func (h *okHandler) Execute(ctx context.Context, q *testQuery) (Response, error) {
	return "ok:" + q.Value, nil
}

type errHandler struct{}

func (h *errHandler) Execute(ctx context.Context, q *testQuery) (Response, error) {
	return nil, errors.New("something went wrong")
}

type handlerQueryA struct{}

func (h *handlerQueryA) Execute(ctx context.Context, q *testQuery) (Response, error) {
	return "A", nil
}

type handlerQueryB struct{}

func (h *handlerQueryB) Execute(ctx context.Context, q *testQuery) (Response, error) {
	return "B", nil
}

func TestCoreBusInMemory_DispatchDeliversToHandler(t *testing.T) {
	bus := NewQueryBus()

	err := bus.Register(&okHandler{})
	assert.NoError(t, err)

	res, err := bus.Dispatch(context.Background(), &testQuery{Value: "x"})
	assert.NoError(t, err)
	assert.Equal(t, "ok:x", res.(string))
}

func TestCoreBusInMemory_DispatchWhenHandlerDoesNotExist(t *testing.T) {
	bus := NewQueryBus()

	_, err := bus.Dispatch(context.Background(), &testQuery{Value: "x"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no handler registered")
}

func TestCoreBusInMemory_RegisterValidations(t *testing.T) {
	bus := NewQueryBus()

	t.Run("Fails when handler is nil", func(t *testing.T) {
		err := bus.Register(nil)
		assert.Error(t, err)
	})

	t.Run("Fails when handler is not a pointer", func(t *testing.T) {
		// okHandler{} es struct, no *struct
		err := bus.Register(okHandler{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be a pointer")
	})

	t.Run("Fails when handler lacks Execute method", func(t *testing.T) {
		type emptyHandler struct{}
		err := bus.Register(&emptyHandler{})
		assert.Error(t, err)
	})
}

func TestCoreBusInMemory_DispatchWhenHandlerReturnsError(t *testing.T) {
	bus := NewQueryBus()
	errReg := bus.Register(&errHandler{})
	assert.NoError(t, errReg)
	res, err := bus.Dispatch(context.Background(), &testQuery{Value: "x"})
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestCoreBusInMemory_RegisterTwiceOverridesHandler(t *testing.T) {
	bus := NewQueryBus()
	errA := bus.Register(&handlerQueryA{})
	assert.NoError(t, errA)

	errB := bus.Register(&handlerQueryB{})
	assert.NoError(t, errB)

	res, err := bus.Dispatch(context.Background(), &testQuery{Value: "x"})
	assert.NoError(t, err)
	assert.Equal(t, "B", res.(string))
}
