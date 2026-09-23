package broadcasterrabbitmqsetup

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-rabbitmq/app/configurationmanager"
	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-rabbitmq/app/sourceprovider"
	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-rabbitmq/domain/broadcaster"
	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-rabbitmq/infrastructure/broadcaster_producer"
	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-rabbitmq/infrastructure/config"
	coreconfigurationmanager "gitlab.com/iobuilders/projects/eng/naryo-go/core/app/configurationmanager"
	corebroadcaster "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
	coreconfig "gitlab.com/iobuilders/projects/eng/naryo-go/core/infrastructure/config"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/infrastructure/event_mapper"
)

type BroadcasterRabbitMQModule struct {
	RabbitMQBroadcasterProducer *broadcasterproducer.RabbitMQBroadcasterProducer
}

func Setup(ctx context.Context, applicationYml string, rootProperty string, broadcasterConfigConfigManager coreconfigurationmanager.BroadcasterConfigurationConfigurationManager) (*BroadcasterRabbitMQModule, error) {
	rabbitMQBroadcaster, err := loadRabbitMQBroadcaster(ctx, applicationYml, rootProperty)
	if err != nil {
		return nil, err
	}

	if err := broadcasterConfigConfigManager.RegisterMapper(
		ctx,
		broadcaster.TypeRabbitMQ.String(),
		func(ctx context.Context, source corebroadcaster.Configuration) (corebroadcaster.Configuration, error) {
			return broadcaster.NewRabbitMQConfiguration(source)
		}); err != nil {
		return nil, err
	}

	rabbitMQBroadcasterProducer, err := broadcasterproducer.NewRabbitMQBroadcasterProducer(rabbitMQBroadcaster, eventmapper.NewEventToJsonMapper())
	if err != nil {
		return nil, err
	}

	return &BroadcasterRabbitMQModule{
		RabbitMQBroadcasterProducer: rabbitMQBroadcasterProducer,
	}, nil
}

func loadRabbitMQBroadcaster(ctx context.Context, applicationYml string, rootProperty string) (*broadcaster.RabbitMQBroadcaster, error) {
	envProperties, err := coreconfig.LoadConfig[config.RabbitMQEnvironmentProperties](ctx, applicationYml, rootProperty)
	if err != nil {
		return nil, err
	}

	envRabbitMQBroadcasterSourceProvider := config.NewEnvRabbitMQBroadcasterSourceProvider(envProperties)
	rabbitMQBroadcasterConfigManager := configurationmanager.NewDefaultRabbitMQBroadcasterConfigurationManager([]sourceprovider.RabbitMQBroadcasterSourceProvider{envRabbitMQBroadcasterSourceProvider})

	return rabbitMQBroadcasterConfigManager.Load(ctx)
}
