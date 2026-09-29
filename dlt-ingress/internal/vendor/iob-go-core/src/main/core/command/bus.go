package command

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"
)

type Command interface {
	CommandName() string
}

type ProcessCommand struct {
	CallbackUrl string
}

type Response interface {
	event.Event
}

type Handler[T Command, R Response] interface {
	Handle(ctx context.Context, command T) (R, error)
}

type Bus interface {
	RegisterHandler(handler any) error
	Dispatch(ctx context.Context, command Command) (Response, error)
}
