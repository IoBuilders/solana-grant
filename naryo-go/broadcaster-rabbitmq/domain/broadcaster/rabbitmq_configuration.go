package broadcaster

import (
	"github.com/go-viper/mapstructure/v2"
	"github.com/google/uuid"

	corebroadcaster "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

type RabbitMQConfiguration struct {
	id       uuid.UUID
	Exchange Exchange
}

type rabbitMQConfigurationAdditionalProps struct {
	Destination struct {
		Exchange string `mapstructure:"exchange"`
	} `mapstructure:"destination"`
}

func NewRabbitMQConfiguration(domain corebroadcaster.Configuration) (*RabbitMQConfiguration, error) {
	var additionalProps rabbitMQConfigurationAdditionalProps
	if err := mapstructure.Decode(domain.AdditionalProperties(), &additionalProps); err != nil {
		return nil, err
	}

	exchange, err := NewExchange(additionalProps.Destination.Exchange)
	if err != nil {
		return nil, err
	}

	c := &RabbitMQConfiguration{
		id:       domain.ID(),
		Exchange: exchange,
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c RabbitMQConfiguration) ID() uuid.UUID {
	return c.id
}

func (c RabbitMQConfiguration) Type() corebroadcaster.Type {
	return TypeRabbitMQ
}

func (c RabbitMQConfiguration) Validate() error {
	if c.id == uuid.Nil {
		return domainerrors.NewEmptyFieldError("ID", "RabbitMQConfiguration")
	}
	if _, err := NewExchange(c.Exchange.String()); err != nil {
		return err
	}
	return nil
}

func (c RabbitMQConfiguration) AdditionalProperties() map[string]interface{} {
	return map[string]interface{}{
		"destination": map[string]interface{}{
			"exchange": c.Exchange.String(),
		},
	}
}

var _ corebroadcaster.Configuration = (*RabbitMQConfiguration)(nil)
