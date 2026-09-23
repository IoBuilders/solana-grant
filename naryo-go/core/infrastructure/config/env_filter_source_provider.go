package config

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/sourceprovider"
)

type EnvFilterSourceProvider struct {
	environmentProperties *EnvironmentProperties
}

func NewEnvFilterSourceProvider(environmentProperties *EnvironmentProperties) *EnvFilterSourceProvider {
	return &EnvFilterSourceProvider{environmentProperties: environmentProperties}
}

func (sp *EnvFilterSourceProvider) Load(ctx context.Context) ([]descriptor.Filter, error) {
	descriptors := make([]descriptor.Filter, 0, len(sp.environmentProperties.Filters))
	for _, filter := range sp.environmentProperties.Filters {
		descriptors = append(descriptors, filter)
	}
	return descriptors, nil
}

// Priority of the source provider
func (sp *EnvFilterSourceProvider) Priority() int {
	return 1
}

var _ sourceprovider.FilterSourceProvider = (*EnvFilterSourceProvider)(nil)
