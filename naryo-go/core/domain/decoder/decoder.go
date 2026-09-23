package decoder

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter"
)

// RawPayload is the source-agnostic input to a Decoder. The ingestion
// adapter locates the bytes for a Specification's AnchorSource (or SPL
// instruction data) and fills this in:
//   - EMIT_LOG:    Data is the base64-decoded bytes of the "Program data:"
//     log line.
//   - EMIT_CPI:    Data is the inner-instruction data (self-CPI sentinel +
//     discriminator + Borsh fields).
//   - INSTRUCTION: Data is the instruction data (discriminator + Borsh
//     args).
//   - SPL_NATIVE:  Data is the instruction data; Accounts are the resolved
//     account addresses.
type RawPayload struct {
	ProgramID string
	Data      []byte
	Accounts  []string
}

// Decoder decodes a RawPayload against a filter Specification. Concrete
// implementations know how to recognize and decode the payload shape a
// given filter.Strategy produces.
type Decoder interface {
	// Decode returns the decoded parameters, or (nil, nil) if raw does not
	// correspond to spec (no match — not an error, and cheap: it must never
	// abort processing the rest of the block).
	Decode(spec filter.Specification, raw RawPayload) ([]parameter.ContractEventParameter, error)
}

// Registry routes a Decode call to the Decoder registered for spec's
// Strategy, keeping the door open for more chains/strategies without
// changing callers.
type Registry map[filter.Strategy]Decoder

// Decode implements Decoder by dispatching to the registered Decoder for
// spec.Strategy(). A Strategy with no registered Decoder is a no-match, not
// an error.
func (r Registry) Decode(spec filter.Specification, raw RawPayload) ([]parameter.ContractEventParameter, error) {
	d, ok := r[spec.Strategy()]
	if !ok {
		return nil, nil
	}
	return d.Decode(spec, raw)
}

var _ Decoder = Registry{}
