package config

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-rabbitmq/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-rabbitmq/app/sourceprovider"
)

type EnvRabbitMQBroadcasterSourceProvider struct {
	environmentProperties *RabbitMQEnvironmentProperties
}

func NewEnvRabbitMQBroadcasterSourceProvider(environmentProperties *RabbitMQEnvironmentProperties) *EnvRabbitMQBroadcasterSourceProvider {
	return &EnvRabbitMQBroadcasterSourceProvider{environmentProperties: environmentProperties}
}

func (sp *EnvRabbitMQBroadcasterSourceProvider) Load(_ context.Context) (descriptor.RabbitMQBroadcaster, error) {
	return sp.environmentProperties.RabbitMQ, nil
}

func (sp *EnvRabbitMQBroadcasterSourceProvider) Priority() int {
	return 1
}

var _ sourceprovider.RabbitMQBroadcasterSourceProvider = (*EnvRabbitMQBroadcasterSourceProvider)(nil)
