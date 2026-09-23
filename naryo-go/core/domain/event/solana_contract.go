package event

import (
	"github.com/google/uuid"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
)

// SolanaContractEvent is the Solana implementation of ContractEvent, matched
// against a filter and carrying the arguments decoded from the emitting
// program's IDL.
type SolanaContractEvent struct {
	nodeID     uuid.UUID
	parameters []parameter.ContractEventParameter
	eventName  string
	status     ContractEventStatus

	ProgramID string
	Signature string
	Slot      uint64
}

func NewSolanaContractEvent(
	nodeID uuid.UUID,
	programID, signature string,
	slot uint64,
	parameters []parameter.ContractEventParameter,
	eventName string,
	status ContractEventStatus,
) (SolanaContractEvent, error) {
	if nodeID == uuid.Nil {
		return SolanaContractEvent{}, domainerrors.NewEmptyFieldError("NodeID", "SolanaContractEvent")
	}
	if programID == "" {
		return SolanaContractEvent{}, domainerrors.NewEmptyFieldError("ProgramID", "SolanaContractEvent")
	}
	if signature == "" {
		return SolanaContractEvent{}, domainerrors.NewEmptyFieldError("Signature", "SolanaContractEvent")
	}
	if eventName == "" {
		return SolanaContractEvent{}, domainerrors.NewEmptyFieldError("EventName", "SolanaContractEvent")
	}
	if !status.IsValid() {
		return SolanaContractEvent{}, domainerrors.NewInvalidFieldError(
			"Status", "SolanaContractEvent", "unsupported status "+status.String(),
		)
	}
	if parameters == nil {
		parameters = []parameter.ContractEventParameter{}
	}
	return SolanaContractEvent{
		nodeID:     nodeID,
		parameters: parameters,
		eventName:  eventName,
		status:     status,
		ProgramID:  programID,
		Signature:  signature,
		Slot:       slot,
	}, nil
}

func (e SolanaContractEvent) EventType() Type {
	return TypeContract
}

func (e SolanaContractEvent) NodeID() uuid.UUID {
	return e.nodeID
}

func (e SolanaContractEvent) Parameters() []parameter.ContractEventParameter {
	return e.parameters
}

func (e SolanaContractEvent) EventName() string {
	return e.eventName
}

func (e SolanaContractEvent) Status() ContractEventStatus {
	return e.status
}

var _ ContractEvent = SolanaContractEvent{}
