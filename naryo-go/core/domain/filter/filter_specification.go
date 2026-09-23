package filter

import "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"

// Specification is the chain-specific description a Filter matches events
// against. Each chain package provides its own concrete implementations (see
// the solana package), since the metadata needed to decode and match an event
// differs per strategy. This interface is the port those implementations
// satisfy, keeping the filter aggregate chain-agnostic.
type Specification interface {
	Strategy() Strategy

	// EventName identifies which on-chain event/instruction this
	// specification decodes, to be copied onto a ContractEvent built from a
	// successful decode (e.g. an Anchor IDL event name, or an SPL
	// instruction name).
	EventName() string

	// Matches reports whether e satisfies this specification (event name and
	// parameters, for strategies that support name/parameter matching).
	Matches(e event.ContractEvent) bool
}
