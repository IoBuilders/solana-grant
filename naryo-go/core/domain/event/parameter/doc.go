// Package parameter defines the decoded arguments carried by a
// event.ContractEvent, typed per the emitting chain's own interface
// description (e.g. Solana's Anchor IDL). Each chain contributes its own
// set of concrete parameter implementations, since the available value
// shapes differ per chain.
package parameter
