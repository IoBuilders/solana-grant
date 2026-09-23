package dispatch

import (
	"context"
	"fmt"
	"sync"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/logging"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/trigger"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

// EventDispatcher is the default Dispatcher. It fans every event out to its
// registered triggers concurrently, isolates the failure — error or panic — of
// any single trigger, and drops expired disposable triggers once a Dispatch
// call has returned.
type EventDispatcher struct {
	mu                 sync.RWMutex
	permanentTriggers  []trigger.PermanentTrigger
	disposableTriggers []trigger.DisposableTrigger
}

// NewEventDispatcher builds an EventDispatcher. Triggers are registered
// afterwards with AddTrigger.
func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{}
}

// AddTrigger registers t. A trigger that satisfies DisposableTrigger is tracked
// as disposable — and thus eligible for removal once expired; every other
// trigger is treated as permanent.
func (d *EventDispatcher) AddTrigger(t trigger.Trigger) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if dt, ok := t.(trigger.DisposableTrigger); ok {
		d.disposableTriggers = append(d.disposableTriggers, dt)
		return
	}
	d.permanentTriggers = append(d.permanentTriggers, t)
}

// Dispatch invokes every registered trigger that supports e, each in its own
// goroutine, so that a slow or panicking trigger can neither block nor break
// others. Errors and panics are logged, never propagated. Expired
// disposable triggers are removed before Dispatch returns.
func (d *EventDispatcher) Dispatch(ctx context.Context, e event.Event) {
	triggers := d.snapshot()

	var wg sync.WaitGroup
	wg.Add(len(triggers))
	for _, t := range triggers {
		go func(t trigger.Trigger) {
			defer wg.Done()
			defer d.recoverTrigger(ctx, t)

			if !t.Supports(e) {
				return
			}
			if err := t.Process(ctx, e); err != nil {
				logging.ErrorWithCtx(
					ctx,
					"dispatch: trigger failed to process event",
					"trigger", triggerName(t),
					"error", err,
				)
			}
		}(t)
	}
	wg.Wait()

	d.RemoveExpired()
}

// RemoveExpired drops every registered disposable trigger whose IsExpired
// reports true.
func (d *EventDispatcher) RemoveExpired() {
	d.mu.Lock()
	defer d.mu.Unlock()

	kept := d.disposableTriggers[:0]
	for _, t := range d.disposableTriggers {
		if t.IsExpired() {
			continue
		}
		kept = append(kept, t)
	}
	// Release the references left in the slice's tail so removed triggers can
	// be garbage-collected.
	for i := len(kept); i < len(d.disposableTriggers); i++ {
		d.disposableTriggers[i] = nil
	}
	d.disposableTriggers = kept
}

// snapshot copies the currently registered triggers into a single slice, so the
// fan-out runs without holding the lock and tolerates concurrent registration
// or removal.
func (d *EventDispatcher) snapshot() []trigger.Trigger {
	d.mu.RLock()
	defer d.mu.RUnlock()

	triggers := make([]trigger.Trigger, 0, len(d.permanentTriggers)+len(d.disposableTriggers))
	for _, t := range d.permanentTriggers {
		triggers = append(triggers, t)
	}
	for _, t := range d.disposableTriggers {
		triggers = append(triggers, t)
	}
	return triggers
}

// recoverTrigger turns a trigger panic into a logged, isolated failure.
func (d *EventDispatcher) recoverTrigger(ctx context.Context, t trigger.Trigger) {
	if r := recover(); r != nil {
		logging.ErrorWithCtx(
			ctx,
			"dispatch: trigger panicked",
			"trigger", triggerName(t),
			"panic", r,
		)
	}
}

func triggerName(t trigger.Trigger) string {
	return fmt.Sprintf("%T", t)
}

var _ Dispatcher = (*EventDispatcher)(nil)
