//go:build test

package common

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func newWsEndpoint(t *testing.T) *ConnectionEndpoint {
	t.Helper()
	endpoint, err := NewConnectionEndpointFromURL("wss://api.mainnet-beta.solana.com")
	assert.NoError(t, err)
	return endpoint
}

func newHTTPEndpoint(t *testing.T) *ConnectionEndpoint {
	t.Helper()
	endpoint, err := NewConnectionEndpointFromURL("https://api.mainnet-beta.solana.com")
	assert.NoError(t, err)
	return endpoint
}

func TestWsConnection_New(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		connection, err := NewWsConnection(newWsEndpoint(t), DefaultRetryConfiguration())
		assert.NoError(t, err)
		assert.Equal(t, ConnectionTypeWs, connection.ConnectionType())
		assert.Equal(t, "wss://api.mainnet-beta.solana.com", connection.Endpoint.URL())
		assert.Equal(t, DefaultRetryConfiguration(), connection.Retry)
	})

	t.Run("HttpEndpoint", func(t *testing.T) {
		_, err := NewWsConnection(newHTTPEndpoint(t), DefaultRetryConfiguration())
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("ZeroEndpoint", func(t *testing.T) {
		_, err := NewWsConnection(&ConnectionEndpoint{}, DefaultRetryConfiguration())
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("ZeroRetry", func(t *testing.T) {
		_, err := NewWsConnection(newWsEndpoint(t), &RetryConfiguration{})
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}

func TestHttpConnection_New(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		connection, err := NewHttpConnection(newHTTPEndpoint(t), DefaultRetryConfiguration())
		assert.NoError(t, err)
		assert.Equal(t, ConnectionTypeHttp, connection.ConnectionType())
		assert.Equal(t, "https://api.mainnet-beta.solana.com", connection.Endpoint.URL())
		assert.Equal(t, DefaultRetryConfiguration(), connection.Retry)
	})

	t.Run("WsEndpoint", func(t *testing.T) {
		_, err := NewHttpConnection(newWsEndpoint(t), DefaultRetryConfiguration())
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("ZeroEndpoint", func(t *testing.T) {
		_, err := NewHttpConnection(&ConnectionEndpoint{}, DefaultRetryConfiguration())
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("ZeroRetry", func(t *testing.T) {
		_, err := NewHttpConnection(newHTTPEndpoint(t), &RetryConfiguration{})
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}
