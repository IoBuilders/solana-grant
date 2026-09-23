package config

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/naryo-go/persistence-gorm/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/persistence-gorm/app/sourceprovider"
)

type EnvDatabaseSourceProvider struct {
	environmentProperties *EnvironmentProperties
}

func NewEnvDatabaseSourceProvider(environmentProperties *EnvironmentProperties) *EnvDatabaseSourceProvider {
	return &EnvDatabaseSourceProvider{environmentProperties: environmentProperties}
}

func (sp *EnvDatabaseSourceProvider) Load(ctx context.Context) (descriptor.Database, error) {
	return sp.environmentProperties.Database, nil
}

// Priority of the source provider
func (sp *EnvDatabaseSourceProvider) Priority() int {
	return 1
}

var _ sourceprovider.DatabaseSourceProvider = (*EnvDatabaseSourceProvider)(nil)
