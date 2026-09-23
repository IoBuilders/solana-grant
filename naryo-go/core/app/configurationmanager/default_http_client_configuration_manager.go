package configurationmanager

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/sourceprovider"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/httpclient"
)

type DefaultHttpClientConfigurationManager struct {
	BaseConfigurationManager[*httpclient.HttpClient, descriptor.HttpClient]
}

func NewDefaultHttpClientConfigurationManager(providers []sourceprovider.HttpClientSourceProvider) *DefaultHttpClientConfigurationManager {
	return &DefaultHttpClientConfigurationManager{
		BaseConfigurationManager: NewBaseConfigurationManager(providers),
	}
}
