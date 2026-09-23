package config

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/sourceprovider"
)

type EnvNodeSourceProvider struct {
	environmentProperties *EnvironmentProperties
}

func NewEnvNodeSourceProvider(environmentProperties *EnvironmentProperties) *EnvNodeSourceProvider {
	return &EnvNodeSourceProvider{environmentProperties: environmentProperties}
}

func (sp *EnvNodeSourceProvider) Load(ctx context.Context) ([]descriptor.Node, error) {
	descriptors := make([]descriptor.Node, 0, len(sp.environmentProperties.Nodes))
	for _, n := range sp.environmentProperties.Nodes {
		descriptors = append(descriptors, n)
	}
	return descriptors, nil
}

// Priority of the source provider
func (sp *EnvNodeSourceProvider) Priority() int {
	return 1
}

var _ sourceprovider.NodeSourceProvider = (*EnvNodeSourceProvider)(nil)
