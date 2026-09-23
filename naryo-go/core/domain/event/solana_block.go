package event

import (
	"github.com/google/uuid"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

// SolanaBlockEvent is the Solana implementation of BlockEvent.
type SolanaBlockEvent struct {
	nodeID       uuid.UUID
	transactions []SolanaTransaction

	Slot      uint64
	Blockhash string
	BlockTime *int64 // nullable, unix seconds
}

func NewSolanaBlockEvent(nodeID uuid.UUID, slot uint64, blockhash string, blockTime *int64, transactions []SolanaTransaction) (SolanaBlockEvent, error) {
	if nodeID == uuid.Nil {
		return SolanaBlockEvent{}, domainerrors.NewEmptyFieldError("NodeID", "SolanaBlockEvent")
	}
	if blockhash == "" {
		return SolanaBlockEvent{}, domainerrors.NewEmptyFieldError("Blockhash", "SolanaBlockEvent")
	}
	if transactions == nil {
		transactions = []SolanaTransaction{}
	}
	return SolanaBlockEvent{
		nodeID:       nodeID,
		transactions: transactions,
		Slot:         slot,
		Blockhash:    blockhash,
		BlockTime:    blockTime,
	}, nil
}

func (b SolanaBlockEvent) EventType() Type {
	return TypeBlock
}

func (b SolanaBlockEvent) NodeID() uuid.UUID {
	return b.nodeID
}

func (b SolanaBlockEvent) Transactions() []BlockTransaction {
	transactions := make([]BlockTransaction, len(b.transactions))
	for i, tx := range b.transactions {
		transactions[i] = tx
	}
	return transactions
}

var _ BlockEvent = SolanaBlockEvent{}
