package store

import (
	"context"

	"github.com/google/uuid"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

// LatestBlockStore persists lastest block event number.
type LatestBlockStore[D event.BlockEvent] interface {
	EventStore[D]

	Get(ctx context.Context, nodeID uuid.UUID) (uint64, error)
}
