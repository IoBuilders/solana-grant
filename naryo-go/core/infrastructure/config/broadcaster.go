package config

import (
	"fmt"

	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster/target"
)

type BroadcastingProperties struct {
	Configuration []*BroadcasterConfigurationProperties `mapstructure:"configuration"`
	Broadcasters  []*BroadcasterProperties              `mapstructure:"broadcasters"`
}

type BroadcasterConfigurationProperties struct {
	ID                   string         `mapstructure:"id"`
	Type                 string         `mapstructure:"type"`
	AdditionalProperties map[string]any `mapstructure:",remain"`
}

func (bc *BroadcasterConfigurationProperties) Map() (broadcaster.Configuration, error) {
	id, err := uuid.Parse(bc.ID)
	if err != nil {
		return nil, err
	}

	return broadcaster.NewGenericConfiguration(id, broadcaster.Type(bc.Type), bc.AdditionalProperties)
}

type BroadcasterEndpointProperties struct {
	URL string `mapstructure:"url"`
}

type BroadcasterProperties struct {
	ID              string            `mapstructure:"id"`
	ConfigurationID string            `mapstructure:"configurationId"`
	Target          BroadcasterTarget `mapstructure:"target"`
}

func (b *BroadcasterProperties) Map() (*broadcaster.Broadcaster, error) {
	id, err := uuid.Parse(b.ID)
	if err != nil {
		return nil, err
	}
	configurationId, err := uuid.Parse(b.ConfigurationID)
	if err != nil {
		return nil, err
	}
	destinations := make([]target.Destination, 0, len(b.Target.Destinations))
	for _, destination := range b.Target.Destinations {
		destinations = append(destinations, target.Destination(destination))
	}
	var trg target.Target
	var targetError error
	switch b.Target.Type {
	case target.TypeAll.String():
		trg, targetError = target.NewAllTarget(destinations)
	case target.TypeBlock.String():
		trg, targetError = target.NewBlockTarget(destinations)
	case target.TypeContractEvent.String():
		trg, targetError = target.NewContractEventTarget(destinations)
	case target.TypeTransaction.String():
		trg, targetError = target.NewTransactionTarget(destinations)
	case target.TypeFilter.String():
		filterID, err := uuid.Parse(b.Target.FilterID)
		if err != nil {
			return nil, err
		}
		trg, targetError = target.NewFilterTarget(destinations, filterID)
	default:
		return nil, fmt.Errorf("unsupported broadcaster target type: %s", b.Target.Type)
	}
	if targetError != nil {
		return nil, targetError
	}
	return broadcaster.NewBroadcaster(id, trg, configurationId)
}

type BroadcasterTarget struct {
	Type         string   `mapstructure:"type"`
	Destinations []string `mapstructure:"destinations"`
	FilterID     string   `mapstructure:"filterId"`
}

var _ descriptor.BroadcasterConfiguration = (*BroadcasterConfigurationProperties)(nil)
var _ descriptor.Broadcaster = (*BroadcasterProperties)(nil)
