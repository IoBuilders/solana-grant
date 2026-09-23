// Package solana provides the Solana-specific implementations of the filter
// domain: the Anchor and native SPL Token specifications, their matching
// strategies, and the Solana data types they decode against. It depends on the
// parent filter package for the Specification port and the Strategy type, and
// filter never depends on it, keeping the aggregate chain-agnostic.
package solana
