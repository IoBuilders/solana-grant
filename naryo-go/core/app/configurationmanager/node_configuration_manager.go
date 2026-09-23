package configurationmanager

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/node"
)

type NodeConfigurationManager interface {
	CollectionConfigurationManager[*node.Node]
}
