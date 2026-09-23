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

func validAdditionalProperties(url string) map[string]interface{} {
	return map[string]interface{}{
		"endpoint": map[string]interface{}{"url": url},
		"retry": map[string]interface{}{
			"maxRetries":   5,
			"initialDelay": "1s",
			"maxDelay":     "30s",
			"multiplier":   2.0,
		},
	}
}

func TestHTTPConfiguration_New(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		id := uuid.New()
		domain, _ := corebroadcaster.NewGenericConfiguration(id, corebroadcaster.TypeHTTP, validAdditionalProperties("https://example.com/webhook"))

		c, err := NewHTTPConfiguration(domain)

		require.NoError(t, err)
		assert.Equal(t, id, c.ID())
		assert.Equal(t, TypeHTTP, c.Type())
		assert.Equal(t, "https://example.com/webhook", c.Connection.Endpoint.URL())
	})

	t.Run("EmptyEndpoint", func(t *testing.T) {
		domain, _ := corebroadcaster.NewGenericConfiguration(uuid.New(), corebroadcaster.TypeHTTP, nil)

		_, err := NewHTTPConfiguration(domain)

		assert.Error(t, err)
	})

	t.Run("InvalidRetry", func(t *testing.T) {
		additionalProperties := map[string]interface{}{
			"endpoint": map[string]interface{}{"url": "https://example.com/webhook"},
			// Retry intentionally left zero-valued (InitialDelay must be > 0)
		}
		domain, _ := corebroadcaster.NewGenericConfiguration(uuid.New(), corebroadcaster.TypeHTTP, additionalProperties)

		_, err := NewHTTPConfiguration(domain)

		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}

func TestHTTPConfiguration_AdditionalProperties_RoundTrips(t *testing.T) {
	domain, _ := corebroadcaster.NewGenericConfiguration(uuid.New(), corebroadcaster.TypeHTTP, validAdditionalProperties("https://example.com/webhook"))
	c, err := NewHTTPConfiguration(domain)
	require.NoError(t, err)

	props := c.AdditionalProperties()

	endpoint, ok := props["endpoint"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "https://example.com/webhook", endpoint["url"])
	retry, ok := props["retry"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, 5, retry["maxRetries"])
}
