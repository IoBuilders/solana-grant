package solana

import "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter"

// Solana matching strategies, as filter.Strategy values. They discriminate how
// a Filter decodes and matches a Solana event: against an Anchor program's
// discriminator layout (ANCHOR) or against the native SPL Token program's
// parsed instructions (SPL_NATIVE). Each chain package contributes its own
// strategy values so the generic filter package stays chain-agnostic.
const (
	StrategyAnchor    filter.Strategy = "ANCHOR"
	StrategySplNative filter.Strategy = "SPL_NATIVE"
)
