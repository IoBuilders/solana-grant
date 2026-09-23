package broadcast

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

// Producer delivers a domain Event to a Broadcaster's target destinations,
// through the transport described by Configuration.
type Producer interface {
	Produce(ctx context.Context, b broadcaster.Broadcaster, configuration broadcaster.Configuration, e event.Event) error
	Supports(t broadcaster.Type) bool
}
