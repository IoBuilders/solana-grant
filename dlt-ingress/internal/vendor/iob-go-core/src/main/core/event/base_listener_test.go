package event

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MyTestEvent struct {
	BaseEvent
	Payload string
}

func TestBaseListener_Listen_Success(t *testing.T) {
	ctx := context.Background()
	event := &MyTestEvent{
		BaseEvent: *NewBaseEvent(),
		Payload:   "test-payload",
	}

	handled := false
	handler := func(ctx context.Context, ev *MyTestEvent) error {
		handled = true
		assert.Equal(t, "test-payload", ev.Payload)
		return nil
	}

	listener := NewBaseListener[MyTestEvent](handler)
	err := listener.Listen(ctx, event)

	assert.NoError(t, err)
	assert.True(t, handled)
}

func TestBaseListener_Listen_WrongType(t *testing.T) {
	ctx := context.Background()
	event := &BaseEvent{
		Id: NewBaseEvent().Id,
	}

	handler := func(ctx context.Context, ev *MyTestEvent) error {
		return nil
	}

	listener := NewBaseListener[MyTestEvent](handler)
	err := listener.Listen(ctx, event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected event type")
}

func TestBaseListener_Listen_NilEvent(t *testing.T) {
	ctx := context.Background()

	handler := func(ctx context.Context, ev *MyTestEvent) error {
		return nil
	}

	listener := NewBaseListener[MyTestEvent](handler)
	err := listener.Listen(ctx, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected event type")
}
