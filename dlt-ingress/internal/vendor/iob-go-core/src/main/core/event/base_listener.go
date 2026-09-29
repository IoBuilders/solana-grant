package event

import (
	"context"
	"fmt"
)

type EventPtr[T any] interface {
	*T
	Event
}

type BaseListener[T any, P EventPtr[T]] struct {
	handle func(context.Context, P) error
}

func NewBaseListener[T any, P EventPtr[T]](h func(context.Context, P) error) *BaseListener[T, P] {
	return &BaseListener[T, P]{handle: h}
}

func (b *BaseListener[T, P]) Listen(ctx context.Context, e Event) error {
	resolved, ok := e.(P)
	if !ok {
		return fmt.Errorf("unexpected event type: expected %T, got %T", *new(P), e)
	}
	return b.handle(ctx, resolved)
}
