package solana

import "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"

// parameterDefinitionBase holds the fields and validation common to every
// ParameterDefinition implementation.
type parameterDefinitionBase struct {
	position int
}

func newParameterDefinitionBase(position int) (parameterDefinitionBase, error) {
	if position < 0 {
		return parameterDefinitionBase{}, domainerrors.NewInvalidFieldError("Position", "ParameterDefinition", "must be non-negative")
	}
	return parameterDefinitionBase{position: position}, nil
}

func (b parameterDefinitionBase) Position() int {
	return b.position
}

// isValidIntegerBitSize mirrors event/parameter's own Uint/Int width
// constraint. Duplicated here deliberately rather than exported from
// event/parameter, to avoid touching an already-shipped, tested file for a
// change entirely local to this package.
func isValidIntegerBitSize(bitSize int) bool {
	switch bitSize {
	case 8, 16, 32, 64, 128:
		return true
	}
	return false
}

// validateNoDuplicatePositions reports an error if any two definitions in
// defs share a Position — used both for AnchorSpecification's top-level
// Parameters and for a StructParameterDefinition's own Fields.
func validateNoDuplicatePositions(defs []ParameterDefinition, entity string) error {
	seen := make(map[int]bool, len(defs))
	for _, d := range defs {
		if seen[d.Position()] {
			return domainerrors.NewInvalidFieldError("Parameters", entity, "duplicate position in parameter definitions")
		}
		seen[d.Position()] = true
	}
	return nil
}
