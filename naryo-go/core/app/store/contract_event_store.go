package store

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

// ContractEventStore persists ContractEvents.
//
// TODO: What identifies one is still open, and now belongs to the model rather than
// this port: EVM uses transaction hash plus log index, while Solana's analogue
// is the instruction index, possibly with the inner instruction index for events
// emitted through CPI.
type ContractEventStore[D event.ContractEvent] interface {
	EventStore[D]
}
