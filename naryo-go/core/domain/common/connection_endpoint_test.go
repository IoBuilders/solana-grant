//go:build test

package common

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func TestConnectionEndpoint_New(t *testing.T) {
	t.Run("ValidSchemes", func(t *testing.T) {
		for _, rawURL := range []string{
			"ws://localhost:8900",
			"wss://api.mainnet-beta.solana.com",
			"http://localhost:8899",
			"https://api.mainnet-beta.solana.com",
		} {
			endpoint, err := NewConnectionEndpointFromURL(rawURL)
			assert.NoError(t, err)
			assert.Equal(t, rawURL, endpoint.URL())
		}
	})

	t.Run("Empty", func(t *testing.T) {
		_, err := NewConnectionEndpointFromURL("")
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("UnsupportedScheme", func(t *testing.T) {
		_, err := NewConnectionEndpointFromURL("ftp://localhost:21")
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("MissingHost", func(t *testing.T) {
		_, err := NewConnectionEndpointFromURL("ws://")
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("Unparseable", func(t *testing.T) {
		_, err := NewConnectionEndpointFromURL("ws://bad url with spaces")
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}

func TestConnectionEndpoint_Scheme(t *testing.T) {
	endpoint, err := NewConnectionEndpointFromURL("wss://api.mainnet-beta.solana.com")
	assert.NoError(t, err)
	assert.Equal(t, "wss", endpoint.Scheme())
}
