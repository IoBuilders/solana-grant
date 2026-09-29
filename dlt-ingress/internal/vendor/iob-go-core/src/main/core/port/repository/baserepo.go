package repository

import (
	"context"

	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api/pagination"
)

type BaseRepository[R any, E any] interface {
	FindById(ctx context.Context, id uuid.UUID) (*E, error)
	FindByIdPreload(ctx context.Context, id uuid.UUID) (*E, error)
	FindAll(context.Context) ([]E, error)
	FindAllWithDeleted(context.Context) ([]E, error)
	FindAllPreload(context.Context) ([]E, error)
	Save(context.Context, *E) error
	Create(context.Context, *E) error
	Delete(context.Context, *E) error
	FindAllPaginated(ctx context.Context, params pagination.PaginationParams) ([]E, error)
	FindAllPaginatedPreload(ctx context.Context, params pagination.PaginationParams) ([]E, error)
	CountAll(context.Context) (int, error)
	ExistById(ctx context.Context, id uuid.UUID) (bool, error)
}
