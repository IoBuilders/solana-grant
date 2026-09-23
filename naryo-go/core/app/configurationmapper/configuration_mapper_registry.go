package configurationmapperregistry

import (
	"context"
	"fmt"
)

type ConfigurationMapperRegistry[S any] interface {
	Register(ctx context.Context, configurationType string, mapper func(ctx context.Context, source S) (S, error)) error
	Map(ctx context.Context, configurationType string, source S) (S, error)
}

type BaseConfigurationMapperRegistry[S any] struct {
	registry map[string]func(ctx context.Context, source S) (S, error)
}

func NewBaseConfigurationMapperRegistry[S any]() *BaseConfigurationMapperRegistry[S] {
	return &BaseConfigurationMapperRegistry[S]{
		registry: make(map[string]func(ctx context.Context, source S) (S, error)),
	}
}

func (r *BaseConfigurationMapperRegistry[S]) Register(_ context.Context, configurationType string, mapper func(ctx context.Context, source S) (S, error)) error {
	r.registry[configurationType] = mapper
	return nil
}

func (r *BaseConfigurationMapperRegistry[S]) Map(ctx context.Context, configurationType string, source S) (S, error) {
	mapper, ok := r.registry[configurationType]
	if !ok {
		var zero S
		return zero, fmt.Errorf("no mapper registered for configuration type %s", configurationType)
	}
	return mapper(ctx, source)
}
