package query

import (
	"context"
	"fmt"
)

type CrossQuery interface {
	CrossQueryName() string
}

type CrossQueryR[R any] interface {
	CrossQuery
	ResponseType() R
}

type CrossResponse interface{}

type CrossHandler[T CrossQuery] interface {
	Execute(ctx context.Context, query CrossQuery) (CrossResponse, error)
}

type CrossBus interface {
	Register(handler any) error
	Dispatch(ctx context.Context, query CrossQuery) (CrossResponse, error)
}

func CrossAsk[R any](ctx context.Context, b CrossBus, q CrossQueryR[R]) (R, error) {
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
