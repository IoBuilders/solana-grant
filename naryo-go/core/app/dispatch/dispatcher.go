package dispatch

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/trigger"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

// Dispatcher fans a domain Event out to every registered Trigger.
type Dispatcher interface {
	// Dispatch invokes every registered Trigger that supports e, isolating the
	// failure of any single Trigger, and removes expired disposable triggers
	// before it returns.
	Dispatch(ctx context.Context, e event.Event)

	// AddTrigger registers a Trigger. Permanent triggers are added at node
	// initialization; disposable triggers may be added at runtime.
	AddTrigger(t trigger.Trigger)

	// RemoveExpired drops every registered DisposableTrigger whose IsExpired
	// reports true.
	RemoveExpired()
}
