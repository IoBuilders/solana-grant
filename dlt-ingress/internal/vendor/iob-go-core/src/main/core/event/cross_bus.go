package event

import (
	"context"
)

type CrossBus interface {
	Publish(ctx context.Context, event CrossEvent) error
}

type CrossEvent interface {
	Event
}
