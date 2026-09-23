package trigger

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/broadcast"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/configurationmanager"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/logging"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/routing"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

// EventBroadcasterPermanentTrigger is the default PermanentTrigger that fans
// an event out to its configured broadcasters. For every event it processes,
// it resolves the Broadcasters routed to it via Router, resolves each
// Broadcaster's Configuration from its own in-memory Configuration snapshot,
// and hands the event to every Producer that supports that Configuration's
// Type. A missing Configuration or a Produce error for one Broadcaster is
// logged and does not stop the others from being processed.
type EventBroadcasterPermanentTrigger struct {
	mu                   sync.RWMutex
	configurations       []broadcaster.Configuration
	router               routing.Router
	producers            []broadcast.Producer
	configurationManager configurationmanager.BroadcasterConfigurationConfigurationManager
}

// NewEventBroadcasterPermanentTrigger builds an EventBroadcasterPermanentTrigger
// backed by router, producers, and configurationManager. Call Reload before
// the first Process to populate the Configuration snapshot.
func NewEventBroadcasterPermanentTrigger(
	router routing.Router,
	producers []broadcast.Producer,
	configurationManager configurationmanager.BroadcasterConfigurationConfigurationManager,
) *EventBroadcasterPermanentTrigger {
	return &EventBroadcasterPermanentTrigger{
		router:               router,
		producers:            producers,
		configurationManager: configurationManager,
	}
}

// Reload replaces the in-memory Configuration snapshot with the current
// output of the BroadcasterConfigurationConfigurationManager.
func (t *EventBroadcasterPermanentTrigger) Reload(ctx context.Context) error {
	configurations, err := t.configurationManager.Load(ctx)
	if err != nil {
		return err
	}

	t.mu.Lock()
	t.configurations = configurations
	t.mu.Unlock()
	return nil
}

// Process routes e through Router, then for every matched Broadcaster
// resolves its Configuration from the trigger's own snapshot and hands e to
// every Producer that supports that Configuration's Type. A missing
// Configuration or a Produce error is logged and the loop continues with the
// next Broadcaster; Process always returns nil, since these are routine
// per-Broadcaster failures, not trigger-level faults.
func (t *EventBroadcasterPermanentTrigger) Process(ctx context.Context, e event.Event) error {
	matched, err := t.router.Match(ctx, e)
	if err != nil {
		logging.ErrorWithCtx(ctx, "event broadcaster trigger: failed to match broadcasters", "error", err)
		return nil
	}

	configurations := t.snapshot()

	for _, b := range matched {
		configuration, ok := configurationFor(configurations, b.ConfigurationID)
		if !ok {
			logging.ErrorWithCtx(
				ctx,
				"event broadcaster trigger: configuration not found for broadcaster",
				"broadcaster_id", b.ID,
				"configuration_id", b.ConfigurationID,
			)
			continue
		}

		for _, producer := range t.producers {
			if !producer.Supports(configuration.Type()) {
				continue
			}
			if err := producer.Produce(ctx, *b, configuration, e); err != nil {
				// TODO: Make it retryable and recoverable to try again
				logging.ErrorWithCtx(
					ctx,
					"event broadcaster trigger: failed to produce event",
					"broadcaster_id", b.ID,
					"error", err,
				)
			}
		}
	}

	return nil
}

// Supports always reports true; matching against the routing configuration
// is Router's responsibility, not this trigger's.
func (t *EventBroadcasterPermanentTrigger) Supports(event.Event) bool {
	return true
}

// snapshot copies out the currently loaded Configurations, so Process runs
// without holding the lock.
func (t *EventBroadcasterPermanentTrigger) snapshot() []broadcaster.Configuration {
	t.mu.RLock()
	defer t.mu.RUnlock()

	configurations := make([]broadcaster.Configuration, len(t.configurations))
	copy(configurations, t.configurations)
	return configurations
}

// configurationFor returns the Configuration in configurations whose ID
// equals configurationID, or false if none matches.
func configurationFor(configurations []broadcaster.Configuration, configurationID uuid.UUID) (broadcaster.Configuration, bool) {
	for _, c := range configurations {
		if c.ID() == configurationID {
			return c, true
		}
	}
	return nil, false
}

var _ PermanentTrigger = (*EventBroadcasterPermanentTrigger)(nil)
