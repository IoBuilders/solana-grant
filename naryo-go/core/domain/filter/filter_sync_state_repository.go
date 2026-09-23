package filter

import (
	"context"

	"github.com/google/uuid"
)

// SyncStateRepository persists the per-Filter SyncState independently of the
// Filter aggregate.
type SyncStateRepository interface {
	Upsert(ctx context.Context, state SyncState) error
	FindByFilterID(ctx context.Context, filterID uuid.UUID) (*SyncState, error)
}
