package config

import (
	"fmt"

	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-kafka/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-kafka/domain/broadcaster"
)

type KafkaBroadcasterProperties struct {
	Brokers []string `mapstructure:"brokers"`
}

func (k *KafkaBroadcasterProperties) Map() (*broadcaster.KafkaBroadcaster, error) {
	if k == nil {
		return nil, fmt.Errorf("kafka module's configuration not provided")
	}
	return broadcaster.NewKafkaBroadcaster(k.Brokers)
}

var _ descriptor.KafkaBroadcaster = (*KafkaBroadcasterProperties)(nil)
