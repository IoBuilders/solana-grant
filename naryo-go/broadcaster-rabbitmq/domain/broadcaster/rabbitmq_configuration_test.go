//go:build test

package broadcaster

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	corebroadcaster "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func validRabbitMQAdditionalProperties(exchange string) map[string]interface{} {
	return map[string]interface{}{
		"destination": map[string]interface{}{"exchange": exchange},
	}
}

func TestRabbitMQConfiguration_New(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		id := uuid.New()
		domain, _ := corebroadcaster.NewGenericConfiguration(id, TypeRabbitMQ, validRabbitMQAdditionalProperties("naryo-events"))

		c, err := NewRabbitMQConfiguration(domain)

		require.NoError(t, err)
		assert.Equal(t, id, c.ID())
		assert.Equal(t, TypeRabbitMQ, c.Type())
		assert.Equal(t, Exchange("naryo-events"), c.Exchange)
	})

	t.Run("MissingDestination", func(t *testing.T) {
		domain, _ := corebroadcaster.NewGenericConfiguration(uuid.New(), TypeRabbitMQ, nil)

		_, err := NewRabbitMQConfiguration(domain)

		assert.ErrorIs(t, err, domainerrors.ErrValidation)
		assert.EqualError(t, err, "Field Exchange of RabbitMQConfiguration cannot be empty")
	})

	t.Run("BlankExchange", func(t *testing.T) {
		domain, _ := corebroadcaster.NewGenericConfiguration(uuid.New(), TypeRabbitMQ, validRabbitMQAdditionalProperties("  "))

		_, err := NewRabbitMQConfiguration(domain)

		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("FlatExchangeIsNotRead", func(t *testing.T) {
		domain, _ := corebroadcaster.NewGenericConfiguration(uuid.New(), TypeRabbitMQ, map[string]interface{}{"exchange": "naryo-events"})

		_, err := NewRabbitMQConfiguration(domain)

		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("MalformedDestination", func(t *testing.T) {
		domain, _ := corebroadcaster.NewGenericConfiguration(uuid.New(), TypeRabbitMQ, map[string]interface{}{"destination": "naryo-events"})

		_, err := NewRabbitMQConfiguration(domain)

		assert.Error(t, err)
	})
}

func TestRabbitMQConfiguration_AdditionalProperties_RoundTrips(t *testing.T) {
	domain, _ := corebroadcaster.NewGenericConfiguration(uuid.New(), TypeRabbitMQ, validRabbitMQAdditionalProperties("naryo-events"))
	c, err := NewRabbitMQConfiguration(domain)
	require.NoError(t, err)

	props := c.AdditionalProperties()

	destination, ok := props["destination"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "naryo-events", destination["exchange"])
}
