package eventmapper

import (
	"github.com/google/uuid"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

type TransactionEventPayload struct {
	NodeID       uuid.UUID            `json:"nodeId"`
	Signature    string               `json:"signature"`
	Slot         uint64               `json:"slot"`
	HasError     bool                 `json:"hasError"`
	Reason       *string              `json:"reason,omitempty"`
	Accounts     []string             `json:"accounts"`
	Logs         []string             `json:"logs"`
	Instructions []InstructionPayload `json:"instructions"`
}

type InstructionPayload struct {
	ProgramID string   `json:"programId"`
	Data      []byte   `json:"data"`
	Accounts  []string `json:"accounts"`
}

func MapTransactionEvent(e event.SolanaTransactionEvent) TransactionEventPayload {
	instructions := make([]InstructionPayload, len(e.Instructions))
	for i, instr := range e.Instructions {
		instructions[i] = InstructionPayload{
			ProgramID: instr.ProgramID,
			Data:      instr.Data,
			Accounts:  instr.Accounts,
		}
	}

	return TransactionEventPayload{
		NodeID:       e.NodeID(),
		Signature:    e.Signature,
		Slot:         e.Slot,
		HasError:     e.HasError(),
		Reason:       e.DecodedReason,
		Accounts:     e.Accounts,
		Logs:         e.Logs,
		Instructions: instructions,
	}
}
