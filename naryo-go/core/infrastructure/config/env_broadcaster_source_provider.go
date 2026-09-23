package config

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/sourceprovider"
)

type EnvBroadcasterSourceProvider struct {
	environmentProperties *EnvironmentProperties
}

func NewEnvBroadcasterSourceProvider(environmentProperties *EnvironmentProperties) *EnvBroadcasterSourceProvider {
	return &EnvBroadcasterSourceProvider{environmentProperties: environmentProperties}
}

func (sp *EnvBroadcasterSourceProvider) Load(ctx context.Context) ([]descriptor.Broadcaster, error) {
	descriptors := make([]descriptor.Broadcaster, 0, len(sp.environmentProperties.Broadcasting.Broadcasters))
	for _, bc := range sp.environmentProperties.Broadcasting.Broadcasters {
		descriptors = append(descriptors, bc)
	}
	return descriptors, nil
}

// Priority of the source provider
func (sp *EnvBroadcasterSourceProvider) Priority() int {
	return 1
}

var _ sourceprovider.BroadcasterSourceProvider = (*EnvBroadcasterSourceProvider)(nil)
