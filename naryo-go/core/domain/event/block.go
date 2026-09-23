package event

// BlockEvent is a protocol-agnostic view of a block ingested from a Node.
// Chain-specific notions of "which block" (e.g. Solana's slot, Ethereum's
// block number) are not part of this contract; they belong on the concrete
// implementation.
type BlockEvent interface {
	Event
	Transactions() []BlockTransaction
}
