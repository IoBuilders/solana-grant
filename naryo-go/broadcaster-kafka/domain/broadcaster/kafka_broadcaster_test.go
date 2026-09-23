//go:build test

package broadcaster

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func TestNewKafkaBroadcaster_Valid(t *testing.T) {
	b, err := NewKafkaBroadcaster([]string{"broker1:9092", "broker2:9092"})

	require.NoError(t, err)
	assert.Equal(t, []string{"broker1:9092", "broker2:9092"}, b.Brokers)
}

func TestNewKafkaBroadcaster_EmptyBrokers(t *testing.T) {
	_, err := NewKafkaBroadcaster(nil)

	assert.ErrorIs(t, err, domainerrors.ErrValidation)
	assert.EqualError(t, err, "Field Brokers of KafkaBroadcaster cannot be empty")
}

func TestNewKafkaBroadcaster_BlankBroker(t *testing.T) {
	_, err := NewKafkaBroadcaster([]string{"broker1:9092", "  "})

	assert.ErrorIs(t, err, domainerrors.ErrValidation)
	assert.EqualError(t, err, "Field Brokers of KafkaBroadcaster is invalid: broker must be a host:port address")
}

func TestNewKafkaBroadcaster_BrokerMissingPort(t *testing.T) {
	_, err := NewKafkaBroadcaster([]string{"broker1"})

	assert.ErrorIs(t, err, domainerrors.ErrValidation)
	assert.EqualError(t, err, "Field Brokers of KafkaBroadcaster is invalid: broker must be a host:port address")
}

func TestNewKafkaBroadcaster_BrokerEmptyPort(t *testing.T) {
	_, err := NewKafkaBroadcaster([]string{"broker1:"})

	assert.ErrorIs(t, err, domainerrors.ErrValidation)
	assert.EqualError(t, err, "Field Brokers of KafkaBroadcaster is invalid: broker must be a host:port address")
}
