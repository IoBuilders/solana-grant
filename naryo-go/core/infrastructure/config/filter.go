package config

import (
	"fmt"

	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter/solana"
)

// FilterProperties is the raw shape of a filter as it appears in the YAML.
// Its fields are a superset across every supported node type/filter type
// combination (Solana, Hedera, EVM...); not every field is populated for
// every filter — which combination is valid is decided by the domain
// mapper, not here. Statuses is raw strings rather than a single enum type
// since TransactionFilter and EventFilter each constrain it to a different
// domain enum (TransactionStatus vs event.ContractEventStatus) — Map picks
// the right one based on Type.
type FilterProperties struct {
	ID              string                        `mapstructure:"id"`
	Name            string                        `mapstructure:"name"`
	NodeID          string                        `mapstructure:"nodeId"`
	Type            string                        `mapstructure:"type"` // e.g. TRANSACTION, EVENT
	IdentifierType  string                        `mapstructure:"identifierType"`
	Value           []string                      `mapstructure:"value"`
	Scope           string                        `mapstructure:"scope"`
	ContractAddress *string                       `mapstructure:"contractAddress"`
	Specification   FilterSpecificationProperties `mapstructure:"specification"`
	Statuses        []string                      `mapstructure:"statuses"`
}

type FilterSpecificationProperties struct {
	Signature      string  `mapstructure:"signature"`
	Source         string  `mapstructure:"source"`
	Strategy       string  `mapstructure:"strategy"`
	Instruction    string  `mapstructure:"instruction"`
	MintAddress    *string `mapstructure:"mintAddress"`
	AccountAddress *string `mapstructure:"accountAddress"`
}

func (fc *FilterProperties) Map() (filter.Filter, error) {
	t := fc.Type
	id, err := uuid.Parse(fc.ID)
	if err != nil {
		return nil, err
	}
	nodeID, err := uuid.Parse(fc.NodeID)
	if err != nil {
		return nil, err
	}
	switch t {
	case filter.FilterTypeEvent.String():
		var spec filter.Specification
		switch fc.Specification.Strategy {
		case solana.StrategyAnchor.String():
			eventName, parameters, parseErr := parseAnchorSignature(fc.Specification.Signature)
			if parseErr != nil {
				return nil, parseErr
			}
			spec, err = solana.NewAnchorSpecification(
				solana.AnchorSource(fc.Specification.Source),
				fc.Specification.AccountAddress,
				eventName,
				parameters,
			)
			if err != nil {
				return nil, err
			}
		case solana.StrategySplNative.String():
			spec, err = solana.NewSplNativeSpecification(
				fc.Specification.Instruction,
				fc.Specification.MintAddress,
			)
			if err != nil {
				return nil, err
			}
		}
		statuses := make([]event.ContractEventStatus, 0, len(fc.Statuses))
		for _, status := range fc.Statuses {
			statuses = append(statuses, event.ContractEventStatus(status))
		}
		return filter.NewEventFilter(id, filter.Name(fc.Name), nodeID, filter.Scope(fc.Scope), fc.ContractAddress, spec, statuses, nil)
	case filter.FilterTypeTransaction.String():
		statuses := make([]filter.TransactionStatus, 0, len(fc.Statuses))
		for _, status := range fc.Statuses {
			statuses = append(statuses, filter.TransactionStatus(status))
		}
		return filter.NewTransactionFilter(id, filter.Name(fc.Name), nodeID, filter.IdentifierType(fc.IdentifierType), fc.Value, statuses)
	default:
		return nil, fmt.Errorf("unsupported filter type: %s", t)
	}
}

var _ descriptor.Filter = (*FilterProperties)(nil)
