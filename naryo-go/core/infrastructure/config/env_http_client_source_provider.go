package config

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/sourceprovider"
)

type EnvHttpClientSourceProvider struct {
	environmentProperties *EnvironmentProperties
}

func NewEnvHttpClientSourceProvider(environmentProperties *EnvironmentProperties) *EnvHttpClientSourceProvider {
	return &EnvHttpClientSourceProvider{environmentProperties: environmentProperties}
}

func (sp *EnvHttpClientSourceProvider) Load(ctx context.Context) (descriptor.HttpClient, error) {
	return sp.environmentProperties.HttpClient, nil
}

// Priority of the source provider
func (sp *EnvHttpClientSourceProvider) Priority() int {
	return 1
}

var _ sourceprovider.HttpClientSourceProvider = (*EnvHttpClientSourceProvider)(nil)
