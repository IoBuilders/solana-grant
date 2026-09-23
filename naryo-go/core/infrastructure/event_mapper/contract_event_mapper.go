package eventmapper

import (
	"github.com/google/uuid"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

type ContractEventPayload struct {
	EventName  string                          `json:"eventName"`
	NodeID     uuid.UUID                       `json:"nodeId"`
	FilterID   *uuid.UUID                      `json:"filterId,omitempty"`
	ProgramID  string                          `json:"programId"`
	Signature  string                          `json:"signature"`
	Slot       uint64                          `json:"slot"`
	Parameters []ContractEventParameterPayload `json:"parameters"`
}

type ContractEventParameterPayload struct {
	Position int    `json:"position"`
	Type     string `json:"type"`
	Value    any    `json:"value"`
}

func MapContractEvent(e event.SolanaContractEvent, filterID *uuid.UUID) ContractEventPayload {
	params := e.Parameters()
	parameters := make([]ContractEventParameterPayload, len(params))
	for i, p := range params {
		parameters[i] = ContractEventParameterPayload{
			Position: p.Position(),
			Type:     p.Type().String(),
			Value:    p.Value(),
		}
	}

	return ContractEventPayload{
		EventName:  e.EventName(),
		NodeID:     e.NodeID(),
		FilterID:   filterID,
		ProgramID:  e.ProgramID,
		Signature:  e.Signature,
		Slot:       e.Slot,
		Parameters: parameters,
	}
}
