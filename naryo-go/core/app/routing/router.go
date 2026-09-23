package routing

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

// Router resolves which Broadcasters should receive a given Event.
type Router interface {
	// Match returns every Broadcaster configured to receive.
	Match(ctx context.Context, e event.Event) ([]*broadcaster.Broadcaster, error)

	// Reload refreshes the Broadcaster snapshot Match is evaluated against.
	Reload(ctx context.Context) error
}
