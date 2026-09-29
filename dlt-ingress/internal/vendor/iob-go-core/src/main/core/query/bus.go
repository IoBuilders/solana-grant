package query

import (
	"context"
	"fmt"
)

type Query interface {
	QueryName() string
}

type QueryR[R any] interface {
	Query
	ResponseType() R
}

type Response any

type Handler[T Query, R any] interface {
	Execute(ctx context.Context, query T) (R, error)
}

type Bus interface {
	Register(handler any) error
	Dispatch(ctx context.Context, q Query) (Response, error)
}

func Ask[R any](ctx context.Context, b Bus, q QueryR[R]) (R, error) {
	var zero R
	rawRes, err := b.Dispatch(ctx, q)
	if err != nil {
		return zero, err
	}

	typedRes, ok := rawRes.(R)
	if !ok {
		return zero, fmt.Errorf("handler returned %T instead of %T", rawRes, zero)
	}
	return typedRes, nil
}
