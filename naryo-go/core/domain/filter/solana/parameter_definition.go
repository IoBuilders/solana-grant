package solana

import "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"

// ParameterDefinition declares the expected shape — type, decode-relevant
// metadata, and position — of one ordered field of an Anchor event or
// instruction. Each parameter.Type has exactly one concrete implementation
// (UintParameterDefinition, ArrayParameterDefinition, ...), mirroring
// parameter.ContractEventParameter's own family on the decoded-output side.
type ParameterDefinition interface {
	Type() parameter.Type
	Position() int
}
