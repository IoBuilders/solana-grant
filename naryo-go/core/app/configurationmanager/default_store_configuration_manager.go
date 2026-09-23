package configurationmanager

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/sourceprovider"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/store"
)

type DefaultStoreConfigurationManager struct {
	BaseCollectionConfigurationManager[store.Configuration, descriptor.Store]
}

func NewDefaultStoreConfigurationManager(providers []sourceprovider.StoreSourceProvider) *DefaultStoreConfigurationManager {
	return &DefaultStoreConfigurationManager{
		BaseCollectionConfigurationManager: NewBaseCollectionConfigurationManager(providers),
	}
}

var _ StoreConfigurationManager = (*DefaultStoreConfigurationManager)(nil)
