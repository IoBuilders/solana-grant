package configurationmanager

import (
	"context"

	configurationmapperregistry "gitlab.com/iobuilders/projects/eng/naryo-go/core/app/configurationmapper"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/sourceprovider"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
)

type DefaultBroadcasterConfigurationConfigurationManager struct {
	BaseCollectionConfigurationManager[broadcaster.Configuration, descriptor.BroadcasterConfiguration]
	mapperRegistry configurationmapperregistry.ConfigurationMapperRegistry[broadcaster.Configuration]
}

func NewDefaultBroadcasterConfigurationConfigurationManager(
	providers []sourceprovider.BroadcasterConfigurationSourceProvider,
	mapperRegistry configurationmapperregistry.ConfigurationMapperRegistry[broadcaster.Configuration]) *DefaultBroadcasterConfigurationConfigurationManager {
	b := &DefaultBroadcasterConfigurationConfigurationManager{
		mapperRegistry: mapperRegistry,
	}
	b.BaseCollectionConfigurationManager = NewBaseCollectionConfigurationManagerWithMapper(providers, b.Map)
	return b
}

func (b *DefaultBroadcasterConfigurationConfigurationManager) Map(ctx context.Context, descriptor descriptor.BroadcasterConfiguration) (broadcaster.Configuration, error) {
	domainBroadcaster, err := descriptor.Map()
	if err != nil {
		return nil, err
	}

	return b.mapperRegistry.Map(ctx, domainBroadcaster.Type().String(), domainBroadcaster)
}

func (b *DefaultBroadcasterConfigurationConfigurationManager) RegisterMapper(ctx context.Context, configurationType string, mapper func(ctx context.Context, source broadcaster.Configuration) (broadcaster.Configuration, error)) error {
	return b.mapperRegistry.Register(ctx, configurationType, mapper)
}
