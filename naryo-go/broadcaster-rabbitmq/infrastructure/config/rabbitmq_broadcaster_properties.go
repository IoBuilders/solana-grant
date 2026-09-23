package config

import (
	"fmt"

	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-rabbitmq/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-rabbitmq/domain/broadcaster"
)

type RabbitMQBroadcasterProperties struct {
	Host        string                `mapstructure:"host"`
	Port        int                   `mapstructure:"port"`
	VirtualHost string                `mapstructure:"virtualHost"`
	Username    string                `mapstructure:"username"`
	Password    string                `mapstructure:"password"`
	TLS         RabbitMQTLSProperties `mapstructure:"tls"`
}

type RabbitMQTLSProperties struct {
	Enabled   bool   `mapstructure:"enabled"`
	Algorithm string `mapstructure:"algorithm"`
}

func (r *RabbitMQBroadcasterProperties) Map() (*broadcaster.RabbitMQBroadcaster, error) {
	if r == nil {
		return nil, fmt.Errorf("rabbitmq module's configuration not provided")
	}
	return broadcaster.NewRabbitMQBroadcaster(
		r.Host,
		r.Port,
		r.VirtualHost,
		r.Username,
		r.Password,
		r.TLS.Enabled,
		broadcaster.TLSAlgorithm(r.TLS.Algorithm),
	)
}

var _ descriptor.RabbitMQBroadcaster = (*RabbitMQBroadcasterProperties)(nil)
