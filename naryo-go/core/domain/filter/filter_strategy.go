package filter

// Strategy discriminates how a Filter decodes and matches an event. It is an
// opaque discriminator: each chain package contributes its own values (e.g.
// the solana package defines ANCHOR and SPL_NATIVE), so this package stays
// chain-agnostic and does not enumerate them.
type Strategy string

func (s Strategy) String() string {
	return string(s)
}
