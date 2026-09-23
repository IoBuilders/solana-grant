package store

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

// TransactionEventStore persists TransactionEvents.
type TransactionEventStore[D event.TransactionEvent] interface {
	EventStore[D]
}
