package broadcaster

import (
	"context"

	"github.com/google/uuid"
)

// ConfigurationRepository persists Configuration implementations.
type ConfigurationRepository interface {
	Save(ctx context.Context, c Configuration) error
	FindByID(ctx context.Context, id uuid.UUID) (Configuration, error)
	FindAll(ctx context.Context) ([]Configuration, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// Repository persists Broadcaster aggregates.
type Repository interface {
	Save(ctx context.Context, b *Broadcaster) error
	FindByFilterID(ctx context.Context, filterID uuid.UUID) ([]Broadcaster, error)
	FindAll(ctx context.Context) ([]Broadcaster, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
