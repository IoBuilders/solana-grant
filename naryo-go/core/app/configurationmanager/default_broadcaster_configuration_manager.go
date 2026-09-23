package configurationmanager

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/sourceprovider"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
)

type DefaultBroadcasterConfigurationManager struct {
	BaseCollectionConfigurationManager[*broadcaster.Broadcaster, descriptor.Broadcaster]
}

func NewDefaultBroadcasterConfigurationManager(providers []sourceprovider.BroadcasterSourceProvider) *DefaultBroadcasterConfigurationManager {
	return &DefaultBroadcasterConfigurationManager{
		BaseCollectionConfigurationManager: NewBaseCollectionConfigurationManager(providers),
	}
}
