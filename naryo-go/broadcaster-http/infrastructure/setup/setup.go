package broadcasterhttpsetup

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-http/domain/broadcaster"
	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-http/infrastructure/http_broadcaster_producer"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/configurationmanager"
	corebroadcaster "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/infrastructure/event_mapper"
)

type BroadcasterHttpModule struct {
	HttpBroadcasterProducer *httpbroadcasterproducer.HttpProducer
}

func Setup(
	ctx context.Context,
	httpConfigManager configurationmanager.HttpClientConfigurationManager,
	broadcasterConfigConfigManager configurationmanager.BroadcasterConfigurationConfigurationManager,
) (*BroadcasterHttpModule, error) {
	httpClient, err := httpConfigManager.Load(ctx)
	if err != nil {
		return nil, err
	}

	if err := broadcasterConfigConfigManager.RegisterMapper(
		ctx,
		broadcaster.TypeHTTP.String(),
		func(ctx context.Context, source corebroadcaster.Configuration) (corebroadcaster.Configuration, error) {
			return broadcaster.NewHTTPConfiguration(source)
		}); err != nil {
		return nil, err
	}

	httpProducer := httpbroadcasterproducer.NewHttpProducer(httpClient, eventmapper.NewEventToJsonMapper())

	return &BroadcasterHttpModule{
		HttpBroadcasterProducer: httpProducer,
	}, nil
}
