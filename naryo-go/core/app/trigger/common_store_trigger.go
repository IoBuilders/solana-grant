package trigger

import (
	"context"

	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/configurationmanager"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/store"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/store/feature"
)

func getStoreFeatureConfigFromConfigManager[T feature.Configuration](ctx context.Context, storeConfigmanager configurationmanager.StoreConfigurationManager, featureType feature.Type, nodeID uuid.UUID) (T, error) {
	var emptyConfig T
	storeConfigs, err := storeConfigmanager.Load(ctx)
	if err != nil {
		return emptyConfig, err
	}
	var activeConfig *store.ActiveConfiguration
	for _, storeConfig := range storeConfigs {
		if storeConfig.NodeID() == nodeID && storeConfig.State() == store.StateActive {
			activeConfig = storeConfig.(*store.ActiveConfiguration)
			break
		}
	}
	if activeConfig == nil {
		return emptyConfig, nil
	}
	featureConfig, okEvent := activeConfig.Feature(featureType)
	if !okEvent {
		return emptyConfig, nil
	}
	return featureConfig.(T), nil
}
