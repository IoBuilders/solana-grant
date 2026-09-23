package configurationmanager

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-rabbitmq/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-rabbitmq/app/sourceprovider"
	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-rabbitmq/domain/broadcaster"
	coreconfigurationmanager "gitlab.com/iobuilders/projects/eng/naryo-go/core/app/configurationmanager"
)

type DefaultRabbitMQBroadcasterConfigurationManager struct {
	coreconfigurationmanager.BaseConfigurationManager[*broadcaster.RabbitMQBroadcaster, descriptor.RabbitMQBroadcaster]
}

func NewDefaultRabbitMQBroadcasterConfigurationManager(providers []sourceprovider.RabbitMQBroadcasterSourceProvider) *DefaultRabbitMQBroadcasterConfigurationManager {
	return &DefaultRabbitMQBroadcasterConfigurationManager{
		BaseConfigurationManager: coreconfigurationmanager.NewBaseConfigurationManager(providers),
	}
}
