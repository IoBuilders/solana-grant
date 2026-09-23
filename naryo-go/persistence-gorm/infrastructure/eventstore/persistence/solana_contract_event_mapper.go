package eventstorepersistence

import (
	"encoding/json"
	"fmt"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

func FromSolanaContractEventDomain(domain event.SolanaContractEvent) (SolanaContractEvent, error) {
	parameters := make([]ContractEventParameter, 0, len(domain.Parameters()))
	for _, parameter := range domain.Parameters() {
		value, err := json.Marshal(parameter.Value())
		if err != nil {
			return SolanaContractEvent{}, fmt.Errorf("error converting from event.SolanaContractEvent: %w", err)
		}
		parameters = append(parameters, ContractEventParameter{
			Type:     string(parameter.Type()),
			Position: parameter.Position(),
			Value:    value,
		})
	}
	return SolanaContractEvent{
		NodeID:     domain.NodeID(),
		EventName:  domain.EventName(),
		Status:     string(domain.Status()),
		ProgramID:  domain.ProgramID,
		Signature:  domain.Signature,
		Slot:       domain.Slot,
		Parameters: parameters,
	}, nil
}
