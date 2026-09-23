package configurationmanager

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
)

type BroadcasterConfigurationManager interface {
	CollectionConfigurationManager[*broadcaster.Broadcaster]
}
