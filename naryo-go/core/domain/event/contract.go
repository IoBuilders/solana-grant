package event

import "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"

// ContractEvent is a protocol-agnostic view of a program/contract event
// matched against a filter, carrying the arguments decoded from the
// emitting program's IDL.
type ContractEvent interface {
	Event
	Parameters() []parameter.ContractEventParameter

	// EventName identifies which program event this is, e.g. an Anchor IDL
	// event name. It is not decoded independently — it is copied in from the
	// EventFilter specification that matched during ingestion.
	EventName() string

	// Status is the Solana confirmation level this event was observed at.
	Status() ContractEventStatus
}
