package coresetup

import (
	"context"
	"reflect"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/broadcast"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/configurationmanager"
	configurationmapperregistry "gitlab.com/iobuilders/projects/eng/naryo-go/core/app/configurationmapper"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/node"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/retry"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/sourceprovider"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/decoder"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter/solana"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/infrastructure/config"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/infrastructure/decoder/solana"
)

type ConfigManagers struct {
	NodeConfigManager                 configurationmanager.NodeConfigurationManager
	BroadcasterConfigManager          configurationmanager.BroadcasterConfigurationManager
	BroadcasterConfigConfigManager    configurationmanager.BroadcasterConfigurationConfigurationManager
	FilterConfigManager               configurationmanager.FilterConfigurationManager
	HttpClientConfigManager           configurationmanager.HttpClientConfigurationManager
	StoreConfigManager                configurationmanager.StoreConfigurationManager
	ProgramErrorRegistryConfigManager configurationmanager.ProgramErrorRegistryConfigurationManager
}

type CoreModule struct {
	Bootstrapper   node.Bootstrapper
	ConfigManagers *ConfigManagers
}

func SetupConfigManagers(ctx context.Context, applicationYml string, rootProperty string) (*ConfigManagers, error) {
	envProperties, err := config.LoadConfig[config.EnvironmentProperties](ctx, applicationYml, rootProperty)
	if err != nil {
		return nil, err
	}

	envNodeSourceProvider := config.NewEnvNodeSourceProvider(envProperties)
	nodeConfigManager := configurationmanager.NewDefaultNodeConfigurationManager([]sourceprovider.NodeSourceProvider{envNodeSourceProvider})

	envBroadcasterSourceProvider := config.NewEnvBroadcasterSourceProvider(envProperties)
	broadcasterConfigManager := configurationmanager.NewDefaultBroadcasterConfigurationManager([]sourceprovider.BroadcasterSourceProvider{envBroadcasterSourceProvider})

	envBroadcasterConfigSourceProvider := config.NewEnvBroadcasterConfigurationSourceProvider(envProperties)
	broadcasterConfigMapperRegistry := configurationmapperregistry.NewBaseConfigurationMapperRegistry[broadcaster.Configuration]()
	broadcasterConfigConfigManager := configurationmanager.NewDefaultBroadcasterConfigurationConfigurationManager([]sourceprovider.BroadcasterConfigurationSourceProvider{envBroadcasterConfigSourceProvider}, broadcasterConfigMapperRegistry)

	envFilterSourceProvider := config.NewEnvFilterSourceProvider(envProperties)
	filterConfigManager := configurationmanager.NewDefaultFilterConfigurationManager([]sourceprovider.FilterSourceProvider{envFilterSourceProvider})

	envHttpClientSourceProvider := config.NewEnvHttpClientSourceProvider(envProperties)
	httpClientConfigManager := configurationmanager.NewDefaultHttpClientConfigurationManager([]sourceprovider.HttpClientSourceProvider{envHttpClientSourceProvider})

	storeSourceProvider := config.NewEnvStoreSourceProvider(envProperties)
	storeConfigManager := configurationmanager.NewDefaultStoreConfigurationManager([]sourceprovider.StoreSourceProvider{storeSourceProvider})

	envProgramErrorRegistrySourceProvider := config.NewEnvProgramErrorRegistrySourceProvider(envProperties)
	programErrorRegistryConfigManager := configurationmanager.NewDefaultProgramErrorRegistryConfigurationManager([]sourceprovider.ProgramErrorRegistrySourceProvider{envProgramErrorRegistrySourceProvider})

	return &ConfigManagers{
		NodeConfigManager:                 nodeConfigManager,
		BroadcasterConfigManager:          broadcasterConfigManager,
		BroadcasterConfigConfigManager:    broadcasterConfigConfigManager,
		FilterConfigManager:               filterConfigManager,
		HttpClientConfigManager:           httpClientConfigManager,
		StoreConfigManager:                storeConfigManager,
		ProgramErrorRegistryConfigManager: programErrorRegistryConfigManager,
	}, nil
}

func Setup(configManagers *ConfigManagers, eventStores map[reflect.Type]any, latestBlockStores map[reflect.Type]any, producers []broadcast.Producer) (*CoreModule, error) {
	d := decoder.Registry{}
	d[solana.StrategyAnchor] = solanadecoder.NewAnchorDecoder()
	d[solana.StrategySplNative] = solanadecoder.SplNativeDecoder{}

	retryer := retry.NewCustomRetryer()

	nodeLifeCycle := node.NewDefaultNodeLifecycle()

	nodeInitializer := node.NewDefaultNodeInitializer(
		configManagers.NodeConfigManager,
		configManagers.BroadcasterConfigManager,
		configManagers.BroadcasterConfigConfigManager,
		configManagers.FilterConfigManager,
		configManagers.HttpClientConfigManager,
		configManagers.StoreConfigManager,
		configManagers.ProgramErrorRegistryConfigManager,
		producers, d, retryer, eventStores, latestBlockStores)

	bootstrapper := node.NewDefaultBootstrapper(nodeInitializer, nodeLifeCycle)

	return &CoreModule{
		ConfigManagers: configManagers,
		Bootstrapper:   bootstrapper,
	}, nil
}
