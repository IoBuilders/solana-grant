package configurationmanager

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/sourceprovider"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter"
)

type DefaultFilterConfigurationManager struct {
	BaseCollectionConfigurationManager[filter.Filter, descriptor.Filter]
}

func NewDefaultFilterConfigurationManager(providers []sourceprovider.FilterSourceProvider) *DefaultFilterConfigurationManager {
	return &DefaultFilterConfigurationManager{
		BaseCollectionConfigurationManager: NewBaseCollectionConfigurationManager(providers),
	}
}
