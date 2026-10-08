package txqueueslot

import (
	"time"

	"github.com/google/uuid"
)

type TxQueueSlot struct {
	Id        uuid.UUID
	CreatedAt time.Time
	NetworkId string
	ExpiresAt time.Time
}

func NewTxQueueSlot(id uuid.UUID, networkId string, expirationTime time.Duration) *TxQueueSlot {
	return &TxQueueSlot{
		Id:        id,
		CreatedAt: time.Now(),
		NetworkId: networkId,
		ExpiresAt: time.Now().Add(expirationTime),
	}
}
