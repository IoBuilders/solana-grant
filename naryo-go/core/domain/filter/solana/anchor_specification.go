package solana

import (
	"sort"
	"strings"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter"
)

// AnchorSpecification matches an Anchor program event or instruction by its
// 8-byte discriminator (derived from EventName) and the ordered layout of the
// parameters that follow it. The Source declares where that discriminator is
// read from on-chain since the decoding route differs per source.
// A specification with no Parameters matches a payload that carries only a
// discriminator.
type AnchorSpecification struct {
	Source AnchorSource
	// AccountAddress restricts matches to events/instructions touching this
	// specific account; nil matches regardless of account (see
	// blockprocessor.matchesAccountAddress). A non-nil value must not be
	// blank — that's the caller stating an address and getting it wrong.
	AccountAddress *string
	Parameters     []ParameterDefinition

	eventName string
}

func NewAnchorSpecification(
	source AnchorSource,
	accountAddress *string,
	eventName string,
	parameters []ParameterDefinition,
) (*AnchorSpecification, error) {
	if !source.IsValid() {
		return nil, domainerrors.NewInvalidFieldError(
			"Source", "AnchorSpecification", "unsupported source "+source.String(),
		)
	}
	if accountAddress != nil && strings.TrimSpace(*accountAddress) == "" {
		return nil, domainerrors.NewEmptyFieldError("AccountAddress", "AnchorSpecification")
	}
	if strings.TrimSpace(eventName) == "" {
		return nil, domainerrors.NewEmptyFieldError("EventName", "AnchorSpecification")
	}
	if err := validateNoDuplicatePositions(parameters, "AnchorSpecification"); err != nil {
		return nil, err
	}
	return &AnchorSpecification{
		Source:         source,
		AccountAddress: accountAddress,
		Parameters:     parameters,
		eventName:      eventName,
	}, nil
}

func (s AnchorSpecification) Strategy() filter.Strategy {
	return StrategyAnchor
}

func (s AnchorSpecification) EventName() string {
	return s.eventName
}

// Matches reports whether e was emitted by this specification's event: its
// name matches, and its decoded parameters agree in count, order and type
// with Parameters. Parameter Indexed/Position ordering beyond Type is not
// otherwise inspected, mirroring how Java Naryo's own matching never
// consulted those attributes either — they exist only to drive decoding.
func (s AnchorSpecification) Matches(e event.ContractEvent) bool {
	if s.eventName != e.EventName() {
		return false
	}
	return parametersMatch(s.Parameters, e.Parameters())
}

func parametersMatch(defs []ParameterDefinition, params []parameter.ContractEventParameter) bool {
	if len(defs) != len(params) {
		return false
	}

	sortedDefs := make([]ParameterDefinition, len(defs))
	copy(sortedDefs, defs)
	sort.Slice(sortedDefs, func(i, j int) bool { return sortedDefs[i].Position() < sortedDefs[j].Position() })

	sortedParams := make([]parameter.ContractEventParameter, len(params))
	copy(sortedParams, params)
	sort.Slice(sortedParams, func(i, j int) bool { return sortedParams[i].Position() < sortedParams[j].Position() })

	for i := range sortedDefs {
		if sortedDefs[i].Type() != sortedParams[i].Type() {
			return false
		}
	}
	return true
}

var _ filter.Specification = AnchorSpecification{}
