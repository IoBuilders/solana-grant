package cache

import "context"

type Port interface {
	Get(ctx context.Context, key string) (any, error)
	Set(ctx context.Context, key string, value any) error
	Delete(ctx context.Context, key string) error
}
