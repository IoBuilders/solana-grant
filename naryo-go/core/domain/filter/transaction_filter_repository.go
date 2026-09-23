package filter

import (
	"context"

	"github.com/google/uuid"
)

// TransactionFilterRepository persists TransactionFilter aggregates for a Node.
type TransactionFilterRepository interface {
	Save(ctx context.Context, f *TransactionFilter) error
	FindByID(ctx context.Context, id uuid.UUID) (*TransactionFilter, error)
	FindAllByNodeID(ctx context.Context, nodeID uuid.UUID) ([]*TransactionFilter, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
