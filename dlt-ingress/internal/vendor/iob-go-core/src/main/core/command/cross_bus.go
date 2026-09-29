package command

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"
)

type CrossCommand interface {
	CrossCommandName() string
}

type CrossProcessCommand struct {
	CallbackUrl string
}

type CrossResponse interface {
	event.Event
}

type CrossPort[T CrossCommand, R CrossResponse] interface {
	Execute(ctx context.Context, command T) (R, error)
}

type CrossBus interface {
	RegisterPort(port any) error
	Dispatch(ctx context.Context, command CrossCommand) (CrossResponse, error)
}
