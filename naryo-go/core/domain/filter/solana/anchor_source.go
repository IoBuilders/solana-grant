package solana

// AnchorSource discriminates where an Anchor event or instruction is read
// from on-chain, since the same 8-byte discriminator can surface through
// three different paths that each require a distinct decoding route:
type AnchorSource string

const (
	AnchorSourceEmitLog     AnchorSource = "EMIT_LOG"
	AnchorSourceEmitCPI     AnchorSource = "EMIT_CPI"
	AnchorSourceInstruction AnchorSource = "INSTRUCTION"
)

func (s AnchorSource) IsValid() bool {
	switch s {
	case AnchorSourceEmitLog, AnchorSourceEmitCPI, AnchorSourceInstruction:
		return true
	}
	return false
}

func (s AnchorSource) String() string {
	return string(s)
}
