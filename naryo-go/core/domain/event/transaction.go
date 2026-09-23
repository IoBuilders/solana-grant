package event

// TransactionEvent is a protocol-agnostic view of a transaction within a
// BlockEvent. Chain-specific notions of "which block" (Solana's slot,
// Ethereum's block number) and chain-specific transaction shape (Solana's
// instructions and runtime logs, Ethereum's calldata) are not part of this
// contract; they belong on the concrete implementation.
type TransactionEvent interface {
	Event
	Id() string
	HasError() bool
}

// BlockTransaction is the protocol-agnostic view of a raw transaction inside
// a block, before any TransactionFilter has matched it — the type
// BlockEvent.Transactions() exposes. Unlike TransactionEvent it is not
// itself an Event: it has not been dispatched or persisted yet.
type BlockTransaction interface {
	Id() string
	HasError() bool
}
