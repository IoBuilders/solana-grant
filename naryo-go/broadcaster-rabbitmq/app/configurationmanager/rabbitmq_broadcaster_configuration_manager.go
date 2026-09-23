package configurationmanager

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-rabbitmq/domain/broadcaster"
	coreconfigurationmanager "gitlab.com/iobuilders/projects/eng/naryo-go/core/app/configurationmanager"
)

type RabbitMQBroadcasterConfigurationManager interface {
	coreconfigurationmanager.ConfigurationManager[*broadcaster.RabbitMQBroadcaster]
}
