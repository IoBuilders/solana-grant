package config

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-kafka/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-kafka/app/sourceprovider"
)

type EnvKafkaBroadcasterSourceProvider struct {
	environmentProperties *KafkaEnvironmentProperties
}

func NewEnvKafkaBroadcasterSourceProvider(environmentProperties *KafkaEnvironmentProperties) *EnvKafkaBroadcasterSourceProvider {
	return &EnvKafkaBroadcasterSourceProvider{environmentProperties: environmentProperties}
}

func (sp *EnvKafkaBroadcasterSourceProvider) Load(_ context.Context) (descriptor.KafkaBroadcaster, error) {
	return sp.environmentProperties.Kafka, nil
}

func (sp *EnvKafkaBroadcasterSourceProvider) Priority() int {
	return 1
}

var _ sourceprovider.KafkaBroadcasterSourceProvider = (*EnvKafkaBroadcasterSourceProvider)(nil)
