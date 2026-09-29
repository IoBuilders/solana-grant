package event

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/domain/eventstore"
)

type Relay interface {
	Relay(ctx context.Context, events *[]eventstore.EventStore) error
	// Wait blocks until all in-flight listener goroutines have finished. Call it
	// from the application shutdown hook after the StoreWorker context has been
	// cancelled, so listeners release their DB connections before the process exits.
	Wait()
}
