package config

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/store"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/store/feature"
)

// StoreProperties is one entry of naryo.stores[]: how a single Node's ingested
// data is persisted.
type StoreProperties struct {
	NodeID   string                    `mapstructure:"nodeId"`
	State    string                    `mapstructure:"state"`
	Type     string                    `mapstructure:"type"`
	Features []*StoreFeatureProperties `mapstructure:"features"`
}

// StoreFeatureProperties is one entry of a store's features[].
type StoreFeatureProperties struct {
	Type        string                   `mapstructure:"type"`
	Strategy    string                   `mapstructure:"strategy"`
	Targets     []*StoreTargetProperties `mapstructure:"targets"`
	Destination string                   `mapstructure:"destination"`
}

type StoreTargetProperties struct {
	Type        string `mapstructure:"type"`
	Destination string `mapstructure:"destination"`
}

func (s *StoreProperties) Map() (store.Configuration, error) {
	nodeID, err := uuid.Parse(s.NodeID)
	if err != nil {
		return nil, fmt.Errorf("store configuration: invalid nodeId %q: %w", s.NodeID, err)
	}

	state := store.State(strings.ToUpper(strings.TrimSpace(s.State)))
	if state == "" {
		state = store.StateActive
	}
	if !state.IsValid() {
		return nil, fmt.Errorf("store configuration: unsupported state %q", s.State)
	}
	if state == store.StateInactive {
		return store.NewInactiveConfiguration(nodeID)
	}

	features, err := s.mapFeatures()
	if err != nil {
		return nil, err
	}

	storeType := store.Type(strings.ToUpper(strings.TrimSpace(s.Type)))
	return store.NewActiveConfiguration(nodeID, storeType, features)
}

func (s *StoreProperties) mapFeatures() (map[feature.Type]feature.Configuration, error) {
	features := make(map[feature.Type]feature.Configuration, len(s.Features))
	for _, properties := range s.Features {
		if properties == nil {
			continue
		}
		featureType, configuration, err := properties.mapFeature()
		if err != nil {
			return nil, err
		}
		if _, duplicated := features[featureType]; duplicated {
			return nil, fmt.Errorf("store configuration: duplicate feature %s", featureType)
		}
		features[featureType] = configuration
	}
	return features, nil
}

func (f *StoreFeatureProperties) mapFeature() (feature.Type, feature.Configuration, error) {
	featureType := feature.Type(strings.ToUpper(strings.TrimSpace(f.Type)))

	switch featureType {
	case feature.TypeEvent:
		configuration, err := f.mapEventFeature()
		return featureType, configuration, err
	case feature.TypeFilterSync:
		configuration, err := feature.NewFilterSyncConfiguration(feature.Destination(f.Destination))
		return featureType, configuration, err
	case feature.TypeLatestBlock:
		configuration, err := feature.NewLatestBlockConfiguration(feature.Destination(f.Destination))
		return featureType, configuration, err
	default:
		return "", nil, fmt.Errorf("store configuration: unsupported feature type %q", f.Type)
	}
}

func (f *StoreFeatureProperties) mapEventFeature() (feature.Configuration, error) {
	// BLOCK_BASED is the only strategy, so an omitted one is not an error.
	strategy := feature.Strategy(strings.ToUpper(strings.TrimSpace(f.Strategy)))
	if strategy == "" {
		strategy = feature.StrategyBlockBased
	}

	targets := make([]feature.EventTarget, 0, len(f.Targets))
	for _, properties := range f.Targets {
		if properties == nil {
			continue
		}
		target, err := feature.NewEventTarget(
			feature.TargetType(strings.ToUpper(strings.TrimSpace(properties.Type))),
			feature.Destination(properties.Destination),
		)
		if err != nil {
			return nil, err
		}
		targets = append(targets, target)
	}

	return feature.NewEventConfiguration(strategy, targets)
}
