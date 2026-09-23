package node

import (
	"context"

	"github.com/google/uuid"
)

// Repository persists Node aggregates.
type Repository interface {
	Save(ctx context.Context, n *Node) error
	FindByID(ctx context.Context, id uuid.UUID) (*Node, error)
	FindAll(ctx context.Context) ([]Node, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
