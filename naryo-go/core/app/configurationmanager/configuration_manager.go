package configurationmanager

import (
	"context"
	"slices"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/sourceprovider"
)

// ConfigurationManager is a port defined by the application layer.
// It exposes a single operation to load all configuration values of type T,
// regardless of how or where they are actually sourced (file, remote store,
// in-memory cache, etc.).
type ConfigurationManager[T any] interface {
	// Load returns value of type T, loading them from the underlying
	// source if they haven't been loaded yet.
	Load(ctx context.Context) (T, error)
}

// Configuration manager for collections
type CollectionConfigurationManager[T any] interface {
	ConfigurationManager[[]T]
}

type BaseCollectionConfigurationManager[T any, S descriptor.Descriptor[T]] struct {
	providers []sourceprovider.CollectionSourceProvider[S]
	mapper    func(ctx context.Context, source S) (T, error)
}

// NewBaseCollectionConfigurationManager builds a manager that maps each
// descriptor via its own Descriptor.Map() method.
func NewBaseCollectionConfigurationManager[T any, S descriptor.Descriptor[T]](providers []sourceprovider.CollectionSourceProvider[S]) BaseCollectionConfigurationManager[T, S] {
	return NewBaseCollectionConfigurationManagerWithMapper(providers, nil)
}

// NewBaseCollectionConfigurationManagerWithMapper builds a manager that maps
// each descriptor via the given mapper instead of Descriptor.Map().
func NewBaseCollectionConfigurationManagerWithMapper[T any, S descriptor.Descriptor[T]](
	providers []sourceprovider.CollectionSourceProvider[S],
	mapper func(ctx context.Context, source S) (T, error),
) BaseCollectionConfigurationManager[T, S] {
	slices.SortFunc(providers, func(a, b sourceprovider.CollectionSourceProvider[S]) int {
		return a.Priority() - b.Priority()
	})
	return BaseCollectionConfigurationManager[T, S]{
		providers: providers,
		mapper:    mapper,
	}
}

func (cm *BaseCollectionConfigurationManager[T, S]) Load(ctx context.Context) ([]T, error) {
	domainList := make([]T, 0)
	for _, provider := range cm.providers {
		descriptorList, err := provider.Load(ctx) //TODO Implement merge function to merge same position of all source providers
		if err != nil {
			return nil, err
		}
		for _, descriptor := range descriptorList {
			var domain T
			var err error

			if cm.mapper == nil {
				domain, err = descriptor.Map()
			} else {
				domain, err = cm.mapper(ctx, descriptor)
			}

			if err != nil {
				return nil, err
			}
			domainList = append(domainList, domain)
		}
	}
	return domainList, nil
}

type BaseConfigurationManager[T any, S descriptor.Descriptor[T]] struct {
	providers []sourceprovider.SourceProvider[S]
}

func NewBaseConfigurationManager[T any, S descriptor.Descriptor[T]](providers []sourceprovider.SourceProvider[S]) BaseConfigurationManager[T, S] {
	return BaseConfigurationManager[T, S]{
		providers: providers,
	}
}

func (cm *BaseConfigurationManager[T, S]) Load(ctx context.Context) (T, error) {
	descriptor, err := cm.providers[0].Load(ctx) //TODO Implement merge function to merge provider list entities
	if err != nil {
		var zero T
		return zero, err
	}
	return descriptor.Map()
}
