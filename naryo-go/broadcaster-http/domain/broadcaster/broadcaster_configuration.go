package broadcaster

import (
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/google/uuid"
	corebroadcaster "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/common"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

type HTTPConfiguration struct {
	id         uuid.UUID
	Connection *common.HttpConnection
}

type httpConfigurationAdditionalProps struct {
	Endpoint struct {
		URL string `mapstructure:"url"`
	} `mapstructure:"endpoint"`
	Retry struct {
		MaxRetries   int           `mapstructure:"maxRetries"`
		InitialDelay time.Duration `mapstructure:"initialDelay"`
		MaxDelay     time.Duration `mapstructure:"maxDelay"`
		Multiplier   float64       `mapstructure:"multiplier"`
	} `mapstructure:"retry"`
}

// NewHTTPConfiguration builds an HTTP Configuration, parsing an HttpConnection
// out of domain.AdditionalProperties() (the "endpoint"/"retry" keys carried
// through from the raw broadcaster configuration).
func NewHTTPConfiguration(domain corebroadcaster.Configuration) (*HTTPConfiguration, error) {
	rawAdditionalProps := domain.AdditionalProperties()

	var additionalProps httpConfigurationAdditionalProps
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.StringToTimeDurationHookFunc(),
		Result:     &additionalProps,
	})
	if err != nil {
		return nil, err
	}
	if err := decoder.Decode(rawAdditionalProps); err != nil {
		return nil, err
	}

	connectionEndpoint, err := common.NewConnectionEndpointFromURL(additionalProps.Endpoint.URL)
	if err != nil {
		return nil, err
	}
	retry, err := common.NewRetryConfiguration(additionalProps.Retry.MaxRetries, additionalProps.Retry.InitialDelay, additionalProps.Retry.MaxDelay, additionalProps.Retry.Multiplier)
	if err != nil {
		return nil, err
	}
	connection, err := common.NewHttpConnection(connectionEndpoint, retry)
	if err != nil {
		return nil, err
	}

	c := &HTTPConfiguration{
		id:         domain.ID(),
		Connection: connection,
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c HTTPConfiguration) ID() uuid.UUID {
	return c.id
}

func (c HTTPConfiguration) Type() corebroadcaster.Type {
	return TypeHTTP
}

func (c HTTPConfiguration) Validate() error {
	if c.id == uuid.Nil {
		return domainerrors.NewEmptyFieldError("ID", "HTTPConfiguration")
	}
	return c.Connection.Validate()
}

func (c HTTPConfiguration) AdditionalProperties() map[string]interface{} {
	var props httpConfigurationAdditionalProps
	props.Endpoint.URL = c.Connection.Endpoint.URL()
	props.Retry.MaxRetries = c.Connection.Retry.MaxRetries
	props.Retry.InitialDelay = c.Connection.Retry.InitialDelay
	props.Retry.MaxDelay = c.Connection.Retry.MaxDelay
	props.Retry.Multiplier = c.Connection.Retry.Multiplier

	additionalProperties := make(map[string]interface{})
	if err := mapstructure.Decode(props, &additionalProperties); err != nil {
		return nil
	}
	return additionalProperties
}

var _ corebroadcaster.Configuration = (*HTTPConfiguration)(nil)
