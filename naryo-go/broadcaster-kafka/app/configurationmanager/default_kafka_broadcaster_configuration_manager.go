package configurationmanager

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-kafka/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-kafka/app/sourceprovider"
	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-kafka/domain/broadcaster"
	coreconfigurationmanager "gitlab.com/iobuilders/projects/eng/naryo-go/core/app/configurationmanager"
)

type DefaultKafkaBroadcasterConfigurationManager struct {
	coreconfigurationmanager.BaseConfigurationManager[*broadcaster.KafkaBroadcaster, descriptor.KafkaBroadcaster]
}

func NewDefaultKafkaBroadcasterConfigurationManager(providers []sourceprovider.KafkaBroadcasterSourceProvider) *DefaultKafkaBroadcasterConfigurationManager {
	return &DefaultKafkaBroadcasterConfigurationManager{
		BaseConfigurationManager: coreconfigurationmanager.NewBaseConfigurationManager(providers),
	}
}
