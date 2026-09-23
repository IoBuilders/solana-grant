package configurationmanager

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-kafka/domain/broadcaster"
	coreconfigurationmanager "gitlab.com/iobuilders/projects/eng/naryo-go/core/app/configurationmanager"
)

type KafkaBroadcasterConfigurationManager interface {
	coreconfigurationmanager.ConfigurationManager[*broadcaster.KafkaBroadcaster]
}
