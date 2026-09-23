package event

import "github.com/google/uuid"

// Event is the contract shared by every domain event produced while
// ingesting chain data: it identifies which kind of event it is and which
// Node emitted it.
type Event interface {
	EventType() Type
	NodeID() uuid.UUID
}
