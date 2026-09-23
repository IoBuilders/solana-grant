package store

import (
	"context"

	"github.com/google/uuid"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

// BlockEventStore persists BlockEvents.
type BlockEventStore[D event.BlockEvent] interface {
	EventStore[D]

	// GetLatest returns the highest block number stored for nodeID — Solana's
	// slot, Ethereum's number — so a restarting Node resumes from there instead
	// of re-scanning.
	//
	// It takes the Node explicitly because two Nodes can share a database, and
	// resuming from another one's blocks would silently skip or replay work.
	GetLatest(ctx context.Context, nodeID uuid.UUID) (uint64, error)
}
