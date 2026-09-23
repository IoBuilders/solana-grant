//go:build test

package config

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvHttpClientSourceProvider_Load_ReturnsHttpClient(t *testing.T) {
	props := &HttpClientProperties{
		MaxIdleConnections: 10,
		KeepAliveDuration:  time.Second,
		ConnectTimeout:     time.Second,
		ReadTimeout:        time.Second,
		WriteTimeout:       time.Second,
		CallTimeout:        time.Second,
		PingInterval:       time.Second,
	}
	sp := NewEnvHttpClientSourceProvider(&EnvironmentProperties{HttpClient: props})

	result, err := sp.Load(context.Background())

	require.NoError(t, err)
	assert.Equal(t, props, result)
}

func TestEnvHttpClientSourceProvider_Load_ReturnsNilWhenNotConfigured(t *testing.T) {
	sp := NewEnvHttpClientSourceProvider(&EnvironmentProperties{HttpClient: nil})

	result, err := sp.Load(context.Background())

	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestEnvHttpClientSourceProvider_Priority(t *testing.T) {
	sp := NewEnvHttpClientSourceProvider(&EnvironmentProperties{})

	assert.Equal(t, 1, sp.Priority())
}
