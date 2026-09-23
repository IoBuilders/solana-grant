package solanadecoder

import (
	"bytes"
	"crypto/sha256"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/decoder"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter/solana"
)

// anchorEventCPISentinel is Anchor's EVENT_IX_TAG_LE: sha256("anchor:event"),
// read as a little-endian u64 and re-encoded to bytes. It is a fixed constant
// (independent of any event's own discriminator) that emit_cpi! prepends to
// the self-CPI instruction data, identifying the inner instruction as an
// event dispatch before the event's own discriminator is even read.
var anchorEventCPISentinel = [8]byte{0xe4, 0x45, 0xa5, 0x2e, 0x51, 0xcb, 0x9a, 0x1d}

// Anchor sighash namespaces: "event" identifies an Anchor event's
// discriminator (used by both EMIT_LOG and EMIT_CPI, since both ultimately
// discriminate the same #[event]-derived struct); "global" identifies a
// direct program instruction call's discriminator.
const (
	anchorEventNamespace       = "event"
	anchorInstructionNamespace = "global"
)

// sighash computes an Anchor discriminator: the first 8 bytes of
// sha256("<namespace>:<name>"), per anchor-lang's own sighash convention.
func sighash(namespace, name string) [8]byte {
	sum := sha256.Sum256([]byte(namespace + ":" + name))
	var disc [8]byte
	copy(disc[:], sum[:8])
	return disc
}

// AnchorDecoder is the Decoder for Anchor-strategy filters: it recognizes an
// Anchor program's discriminator in a RawPayload and decodes the remaining
// bytes via DecodeBuffer. It covers all three AnchorSource routes an
// AnchorSpecification can declare — EMIT_LOG and EMIT_CPI (an emitted event,
// read from a log line or a self-CPI call respectively) and INSTRUCTION (a
// direct program instruction call).
type AnchorDecoder struct{}

// NewAnchorDecoder builds an AnchorDecoder.
func NewAnchorDecoder() AnchorDecoder {
	return AnchorDecoder{}
}

// Decode implements decoder.Decoder. It returns (nil, nil) — not an error —
// whenever raw does not correspond to spec: spec is not an
// *solana.AnchorSpecification, raw.Data is too short to carry the expected
// sentinel/discriminator, or the discriminator does not match. A decode error
// is only ever returned once the discriminator has matched and DecodeBuffer
// itself fails on the remaining bytes.
func (AnchorDecoder) Decode(spec filter.Specification, raw decoder.RawPayload) ([]parameter.ContractEventParameter, error) {
	anchorSpec, ok := spec.(*solana.AnchorSpecification)
	if !ok {
		return nil, nil
	}

	data := raw.Data
	var namespace string

	switch anchorSpec.Source {
	case solana.AnchorSourceEmitCPI:
		if len(data) < len(anchorEventCPISentinel) || !bytes.Equal(data[:len(anchorEventCPISentinel)], anchorEventCPISentinel[:]) {
			return nil, nil
		}
		data = data[len(anchorEventCPISentinel):]
		namespace = anchorEventNamespace
	case solana.AnchorSourceEmitLog:
		namespace = anchorEventNamespace
	case solana.AnchorSourceInstruction:
		namespace = anchorInstructionNamespace
	default:
		// Unreachable: Source is validated by NewAnchorSpecification.
		return nil, nil
	}

	discriminator := sighash(namespace, anchorSpec.EventName())
	if len(data) < len(discriminator) || !bytes.Equal(data[:len(discriminator)], discriminator[:]) {
		return nil, nil
	}

	return DecodeBuffer(data[len(discriminator):], anchorSpec.Parameters)
}

var _ decoder.Decoder = AnchorDecoder{}
