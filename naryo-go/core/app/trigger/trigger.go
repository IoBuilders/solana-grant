package trigger

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

// Trigger consumes an Event that a Dispatcher has fanned out. Supports lets a
// Trigger declare, per event, whether it wants to handle it, so the Dispatcher
// can skip triggers the event does not concern.
type Trigger interface {
	// Process handles the event. A returned error is isolated by the
	// Dispatcher (logged, never propagated to sibling triggers).
	Process(ctx context.Context, e event.Event) error

	// Supports reports whether this Trigger wants to handle e.
	Supports(e event.Event) bool
}

// PermanentTrigger is a Trigger that stays registered and is invoked for every
// supported event (e.g. the broadcaster fanout or the event-store writer). It
// is a marker: it adds no methods over Trigger.
type PermanentTrigger interface {
	Trigger
}

// DisposableTrigger is a one-shot Trigger, such as an event-confirmation
// watcher. Once IsExpired reports true, the Dispatcher drops it after the
// current Dispatch call returns.
type DisposableTrigger interface {
	Trigger

	// IsExpired reports whether this Trigger has fulfilled its purpose and
	// must be removed from the Dispatcher.
	IsExpired() bool
}
