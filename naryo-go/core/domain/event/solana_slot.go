package event

import (
	"time"

	"github.com/google/uuid"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

// SlotEvent mirrors a Solana slotSubscribe notification.
type SlotEvent struct {
	nodeID uuid.UUID

	Slot       uint64
	ReceivedAt time.Time
}

func NewSlotEvent(nodeID uuid.UUID, slot uint64, receivedAt time.Time) (SlotEvent, error) {
	if nodeID == uuid.Nil {
		return SlotEvent{}, domainerrors.NewEmptyFieldError("NodeID", "SlotEvent")
	}
	return SlotEvent{
		nodeID:     nodeID,
		Slot:       slot,
		ReceivedAt: receivedAt,
	}, nil
}

func (e SlotEvent) EventType() Type {
	return TypeSlot
}

func (e SlotEvent) NodeID() uuid.UUID {
	return e.nodeID
}

var _ Event = SlotEvent{}
