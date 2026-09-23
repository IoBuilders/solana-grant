// Package filter defines the filter aggregates and their value objects. A
// Filter comes in two kinds behind a sealed Filter port: an EventFilter, which
// matches a Node's ingested events (with a scope, a chain-specific
// Specification and its synchronization state), and a TransactionFilter, which
// matches transactions — chiefly failed ones — by identifier and status. The
// kinds are modeled separately so neither carries fields that are meaningless
// for the other. The package also declares the repository ports that
// persistence adapters implement.
package filter
