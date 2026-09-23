package configurationmanager

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/sourceprovider"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/node"
)

type DefaultNodeConfigurationManager struct {
	BaseCollectionConfigurationManager[*node.Node, descriptor.Node]
}

func NewDefaultNodeConfigurationManager(providers []sourceprovider.NodeSourceProvider) *DefaultNodeConfigurationManager {
	return &DefaultNodeConfigurationManager{
		BaseCollectionConfigurationManager: NewBaseCollectionConfigurationManager(providers),
	}
}
