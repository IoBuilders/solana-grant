package store

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

// EventStore persists a domain Event.
type EventStore[E event.Event] interface {
	Store[E]
}
