package eventmapper

import (
	"github.com/google/uuid"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

type BlockEventPayload struct {
	NodeID                uuid.UUID `json:"nodeId"`
	Slot                  uint64    `json:"slot"`
	Blockhash             string    `json:"blockhash"`
	BlockTime             *int64    `json:"blockTime,omitempty"`
	TransactionSignatures []string  `json:"transactionSignatures"`
}

func MapBlockEvent(e event.SolanaBlockEvent) BlockEventPayload {
	transactions := e.Transactions()
	signatures := make([]string, len(transactions))
	for i, tx := range transactions {
		signatures[i] = tx.Id()
	}

	return BlockEventPayload{
		NodeID:                e.NodeID(),
		Slot:                  e.Slot,
		Blockhash:             e.Blockhash,
		BlockTime:             e.BlockTime,
		TransactionSignatures: signatures,
	}
}
