//go:build test

package broadcaster

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func TestNewRabbitMQBroadcaster_Valid(t *testing.T) {
	b, err := NewRabbitMQBroadcaster("localhost", 5672, "/naryo", "guest", "guest", false, "")

	require.NoError(t, err)
	assert.Equal(t, "localhost", b.Host)
	assert.Equal(t, 5672, b.Port)
	assert.Equal(t, "/naryo", b.VirtualHost)
	assert.Equal(t, "guest", b.Username)
	assert.Equal(t, "guest", b.Password)
	assert.False(t, b.TLSEnabled)
	assert.Empty(t, b.TLSAlgorithm)
}

func TestNewRabbitMQBroadcaster_BlankVirtualHostDefaults(t *testing.T) {
	b, err := NewRabbitMQBroadcaster("localhost", 5672, "   ", "guest", "guest", false, "")

	require.NoError(t, err)
	assert.Equal(t, "/", b.VirtualHost)
}

func TestNewRabbitMQBroadcaster_TLSEnabledDefaultsAlgorithm(t *testing.T) {
	b, err := NewRabbitMQBroadcaster("localhost", 5671, "/", "guest", "guest", true, "  ")

	require.NoError(t, err)
	assert.True(t, b.TLSEnabled)
	assert.Equal(t, TLSAlgorithmTLS12, b.TLSAlgorithm)
}

func TestNewRabbitMQBroadcaster_TLSEnabledKeepsExplicitAlgorithm(t *testing.T) {
	b, err := NewRabbitMQBroadcaster("localhost", 5671, "/", "guest", "guest", true, "TLSv1.3")

	require.NoError(t, err)
	assert.Equal(t, TLSAlgorithmTLS13, b.TLSAlgorithm)
}

func TestNewRabbitMQBroadcaster_TLSDisabledDoesNotDefaultAlgorithm(t *testing.T) {
	b, err := NewRabbitMQBroadcaster("localhost", 5672, "/", "guest", "guest", false, "")

	require.NoError(t, err)
	assert.Empty(t, b.TLSAlgorithm)
}

func TestNewRabbitMQBroadcaster_EmptyHost(t *testing.T) {
	_, err := NewRabbitMQBroadcaster("", 5672, "/", "guest", "guest", false, "")

	assert.ErrorIs(t, err, domainerrors.ErrValidation)
	assert.EqualError(t, err, "Field Host of RabbitMQBroadcaster cannot be empty")
}

func TestNewRabbitMQBroadcaster_BlankHost(t *testing.T) {
	_, err := NewRabbitMQBroadcaster("   ", 5672, "/", "guest", "guest", false, "")

	assert.ErrorIs(t, err, domainerrors.ErrValidation)
	assert.EqualError(t, err, "Field Host of RabbitMQBroadcaster cannot be empty")
}

func TestNewRabbitMQBroadcaster_ZeroPort(t *testing.T) {
	_, err := NewRabbitMQBroadcaster("localhost", 0, "/", "guest", "guest", false, "")

	assert.ErrorIs(t, err, domainerrors.ErrValidation)
	assert.EqualError(t, err, "Field Port of RabbitMQBroadcaster is invalid: port must be between 1 and 65535")
}

func TestNewRabbitMQBroadcaster_NegativePort(t *testing.T) {
	_, err := NewRabbitMQBroadcaster("localhost", -1, "/", "guest", "guest", false, "")

	assert.ErrorIs(t, err, domainerrors.ErrValidation)
	assert.EqualError(t, err, "Field Port of RabbitMQBroadcaster is invalid: port must be between 1 and 65535")
}

func TestNewRabbitMQBroadcaster_PortOutOfRange(t *testing.T) {
	_, err := NewRabbitMQBroadcaster("localhost", 65536, "/", "guest", "guest", false, "")

	assert.ErrorIs(t, err, domainerrors.ErrValidation)
	assert.EqualError(t, err, "Field Port of RabbitMQBroadcaster is invalid: port must be between 1 and 65535")
}

func TestNewRabbitMQBroadcaster_EmptyUsername(t *testing.T) {
	_, err := NewRabbitMQBroadcaster("localhost", 5672, "/", "  ", "guest", false, "")

	assert.ErrorIs(t, err, domainerrors.ErrValidation)
	assert.EqualError(t, err, "Field Username of RabbitMQBroadcaster cannot be empty")
}

func TestNewRabbitMQBroadcaster_EmptyPassword(t *testing.T) {
	_, err := NewRabbitMQBroadcaster("localhost", 5672, "/", "guest", "", false, "")

	assert.ErrorIs(t, err, domainerrors.ErrValidation)
	assert.EqualError(t, err, "Field Password of RabbitMQBroadcaster cannot be empty")
}

func TestNewRabbitMQBroadcaster_TLSDisabledWithAlgorithm(t *testing.T) {
	_, err := NewRabbitMQBroadcaster("localhost", 5672, "/", "guest", "guest", false, "TLSv1.3")

	assert.ErrorIs(t, err, domainerrors.ErrValidation)
	assert.EqualError(t, err, "Field TLSAlgorithm of RabbitMQBroadcaster is invalid: TLS algorithm must not be set while TLS is disabled")
}

func TestNewRabbitMQBroadcaster_TLSDisabledWithBlankAlgorithmIsAccepted(t *testing.T) {
	b, err := NewRabbitMQBroadcaster("localhost", 5672, "/", "guest", "guest", false, "   ")

	require.NoError(t, err)
	assert.False(t, b.TLSEnabled)
}

func TestNewRabbitMQBroadcaster_TLSEnabledWithUnsupportedAlgorithm(t *testing.T) {
	_, err := NewRabbitMQBroadcaster("localhost", 5671, "/", "guest", "guest", true, "TLSv1.1")

	assert.ErrorIs(t, err, domainerrors.ErrValidation)
	assert.EqualError(t, err, `Field TLSAlgorithm of RabbitMQBroadcaster is invalid: unsupported TLS algorithm "TLSv1.1", must be one of TLSv1.2, TLSv1.3`)
}

func TestNewRabbitMQBroadcaster_TLSEnabledWithTLS13(t *testing.T) {
	b, err := NewRabbitMQBroadcaster("localhost", 5671, "/", "guest", "guest", true, TLSAlgorithmTLS13)

	require.NoError(t, err)
	assert.True(t, b.TLSEnabled)
	assert.Equal(t, TLSAlgorithmTLS13, b.TLSAlgorithm)
}
