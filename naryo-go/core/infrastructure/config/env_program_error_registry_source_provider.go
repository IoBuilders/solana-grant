package config

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/sourceprovider"
)

type EnvProgramErrorRegistrySourceProvider struct {
	environmentProperties *EnvironmentProperties
}

func NewEnvProgramErrorRegistrySourceProvider(environmentProperties *EnvironmentProperties) *EnvProgramErrorRegistrySourceProvider {
	return &EnvProgramErrorRegistrySourceProvider{environmentProperties: environmentProperties}
}

func (sp *EnvProgramErrorRegistrySourceProvider) Load(ctx context.Context) (descriptor.ProgramErrorRegistry, error) {
	return sp.environmentProperties.ProgramErrors, nil
}

// Priority of the source provider
func (sp *EnvProgramErrorRegistrySourceProvider) Priority() int {
	return 1
}

var _ sourceprovider.ProgramErrorRegistrySourceProvider = (*EnvProgramErrorRegistrySourceProvider)(nil)
