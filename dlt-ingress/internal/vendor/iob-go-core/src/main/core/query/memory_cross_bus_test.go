package query

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type tesCrossQuery struct {
	Value string
}

func (q *tesCrossQuery) CrossQueryName() string {
	t := reflect.TypeOf(*q)
	return fmt.Sprintf("%s.%s", t.PkgPath(), t.Name())
}

type okCrossQueryHandler struct{}

func (h *okCrossQueryHandler) Execute(ctx context.Context, q *tesCrossQuery) (Response, error) {
	return "ok:" + q.Value, nil
}

type errCrossQueryHandler struct{}

func (h *errCrossQueryHandler) Execute(ctx context.Context, q *tesCrossQuery) (Response, error) {
	return nil, errors.New("something went wrong")
}

type handlerCrossQueryA struct{}

func (h *handlerCrossQueryA) Execute(ctx context.Context, q *tesCrossQuery) (Response, error) {
	return "A", nil
}

type handlerCrossQueryB struct{}

func (h *handlerCrossQueryB) Execute(ctx context.Context, q *tesCrossQuery) (Response, error) {
	return "B", nil
}

func TestCoreCrossBusInMemory_DispatchDeliversToHandler(t *testing.T) {
	bus := NewCrossQueryBus()

	err := bus.Register(&okCrossQueryHandler{})
	assert.NoError(t, err)

	res, err := bus.Dispatch(context.Background(), &tesCrossQuery{Value: "x"})
	assert.NoError(t, err)
	assert.Equal(t, "ok:x", res.(string))
}

func TestCoreCrossBusInMemory_DispatchWhenHandlerDoesNotExist(t *testing.T) {
	bus := NewCrossQueryBus()

	_, err := bus.Dispatch(context.Background(), &tesCrossQuery{Value: "x"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no handler registered")
}

func TestCoreCrossBusInMemory_RegisterValidations(t *testing.T) {
	bus := NewCrossQueryBus()

	t.Run("Fails when handler is nil", func(t *testing.T) {
		err := bus.Register(nil)
		assert.Error(t, err)
	})

	t.Run("Fails when handler is not a pointer", func(t *testing.T) {
		// okHandler{} es struct, no *struct
		err := bus.Register(okCrossQueryHandler{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be a pointer")
	})

	t.Run("Fails when handler lacks Execute method", func(t *testing.T) {
		type emptyHandler struct{}
		err := bus.Register(&emptyHandler{})
		assert.Error(t, err)
	})
}

func TestCoreCrossBusInMemory_DispatchWhenHandlerReturnsError(t *testing.T) {
	bus := NewCrossQueryBus()
	errReg := bus.Register(&errCrossQueryHandler{})
	assert.NoError(t, errReg)
	res, err := bus.Dispatch(context.Background(), &tesCrossQuery{Value: "x"})
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestCoreCrossBusInMemory_RegisterTwiceOverridesHandler(t *testing.T) {
	bus := NewCrossQueryBus()
	errA := bus.Register(&handlerCrossQueryA{})
	assert.NoError(t, errA)

	errB := bus.Register(&handlerCrossQueryB{})
	assert.NoError(t, errB)

	res, err := bus.Dispatch(context.Background(), &tesCrossQuery{Value: "x"})
	assert.NoError(t, err)
	assert.Equal(t, "B", res.(string))
}
