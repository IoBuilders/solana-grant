package broadcasterkafkasetup

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-kafka/app/configurationmanager"
	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-kafka/app/sourceprovider"
	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-kafka/domain/broadcaster"
	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-kafka/infrastructure/broadcaster_producer"
	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-kafka/infrastructure/config"
	coreconfigurationmanager "gitlab.com/iobuilders/projects/eng/naryo-go/core/app/configurationmanager"
	corebroadcaster "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/infrastructure/event_mapper"

	coreconfig "gitlab.com/iobuilders/projects/eng/naryo-go/core/infrastructure/config"
)

type BroadcasterKafkaModule struct {
	KafkaBroadcasterProducer *broadcasterproducer.KafkaBroadcasterProducer
}

func Setup(ctx context.Context, applicationYml string, rootProperty string, broadcasterConfigConfigManager coreconfigurationmanager.BroadcasterConfigurationConfigurationManager) (*BroadcasterKafkaModule, error) {
	kafkaBroadcaster, err := loadKafkaBroadcaster(ctx, applicationYml, rootProperty)
	if err != nil {
		return nil, err
	}

	if err := broadcasterConfigConfigManager.RegisterMapper(
		ctx,
		broadcaster.TypeKafka.String(),
		func(ctx context.Context, source corebroadcaster.Configuration) (corebroadcaster.Configuration, error) {
			return source, nil
		}); err != nil {
		return nil, err
	}

	kafkaBroadcasterProducer, err := broadcasterproducer.NewKafkaBroadcasterProducer(kafkaBroadcaster, eventmapper.NewEventToJsonMapper())
	if err != nil {
		return nil, err
	}

	return &BroadcasterKafkaModule{
		KafkaBroadcasterProducer: kafkaBroadcasterProducer,
	}, nil
}

func loadKafkaBroadcaster(ctx context.Context, applicationYml string, rootProperty string) (*broadcaster.KafkaBroadcaster, error) {
	envProperties, err := coreconfig.LoadConfig[config.KafkaEnvironmentProperties](ctx, applicationYml, rootProperty)
	if err != nil {
		return nil, err
	}

	envKafkaBroadcasterSourceProvider := config.NewEnvKafkaBroadcasterSourceProvider(envProperties)
	kafkaBroadcasterConfigManager := configurationmanager.NewDefaultKafkaBroadcasterConfigurationManager([]sourceprovider.KafkaBroadcasterSourceProvider{envKafkaBroadcasterSourceProvider})

	return kafkaBroadcasterConfigManager.Load(ctx)
}
