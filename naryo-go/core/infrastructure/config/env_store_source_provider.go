package config

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/sourceprovider"
)

type EnvStoreSourceProvider struct {
	environmentProperties *EnvironmentProperties
}

func NewEnvStoreSourceProvider(environmentProperties *EnvironmentProperties) *EnvStoreSourceProvider {
	return &EnvStoreSourceProvider{environmentProperties: environmentProperties}
}

func (sp *EnvStoreSourceProvider) Load(_ context.Context) ([]descriptor.Store, error) {
	stores := sp.environmentProperties.Stores
	descriptors := make([]descriptor.Store, 0, len(stores))
	for _, properties := range stores {
		if properties == nil {
			continue
		}
		descriptors = append(descriptors, properties)
	}
	return descriptors, nil
}

// Priority of the source provider
func (sp *EnvStoreSourceProvider) Priority() int {
	return 1
}

var _ sourceprovider.StoreSourceProvider = (*EnvStoreSourceProvider)(nil)
