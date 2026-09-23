package parameter

// ContractEventParameter is a single decoded argument of a contract event.
// Each chain provides its own set of concrete implementations (e.g.
// Solana's Solana*Parameter types), since the available value shapes and
// metadata differ per chain.
type ContractEventParameter interface {
	Type() Type
	Position() int
	Value() any
}
