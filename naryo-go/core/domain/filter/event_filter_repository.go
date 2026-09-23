package filter

import (
	"context"

	"github.com/google/uuid"
)

// EventFilterRepository persists EventFilter aggregates for a Node.
type EventFilterRepository interface {
	Save(ctx context.Context, f *EventFilter) error
	FindByID(ctx context.Context, id uuid.UUID) (*EventFilter, error)
	FindAllByNodeID(ctx context.Context, nodeID uuid.UUID) ([]*EventFilter, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
