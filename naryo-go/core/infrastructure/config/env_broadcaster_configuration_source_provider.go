package config

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/sourceprovider"
)

type EnvBroadcasterConfigurationSourceProvider struct {
	environmentProperties *EnvironmentProperties
}

func NewEnvBroadcasterConfigurationSourceProvider(environmentProperties *EnvironmentProperties) *EnvBroadcasterConfigurationSourceProvider {
	return &EnvBroadcasterConfigurationSourceProvider{environmentProperties: environmentProperties}
}

func (sp *EnvBroadcasterConfigurationSourceProvider) Load(ctx context.Context) ([]descriptor.BroadcasterConfiguration, error) {
	descriptors := make([]descriptor.BroadcasterConfiguration, 0, len(sp.environmentProperties.Broadcasting.Configuration))
	for _, bc := range sp.environmentProperties.Broadcasting.Configuration {
		descriptors = append(descriptors, bc)
	}
	return descriptors, nil
}

// Priority of the source provider
func (sp *EnvBroadcasterConfigurationSourceProvider) Priority() int {
	return 1
}

var _ sourceprovider.BroadcasterConfigurationSourceProvider = (*EnvBroadcasterConfigurationSourceProvider)(nil)
