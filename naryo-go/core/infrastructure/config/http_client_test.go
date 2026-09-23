//go:build test

package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHttpClientProperties_Map_MapsAllFields(t *testing.T) {
	props := &HttpClientProperties{
		MaxIdleConnections:       10,
		KeepAliveDuration:        30 * time.Second,
		ConnectTimeout:           5 * time.Second,
		ReadTimeout:              10 * time.Second,
		WriteTimeout:             10 * time.Second,
		CallTimeout:              15 * time.Second,
		PingInterval:             20 * time.Second,
		RetryOnConnectionFailure: true,
	}

	result, err := props.Map()

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, props.MaxIdleConnections, result.MaxIdleConnections)
	assert.Equal(t, props.KeepAliveDuration, result.KeepAliveDuration)
	assert.Equal(t, props.ConnectTimeout, result.ConnectTimeout)
	assert.Equal(t, props.ReadTimeout, result.ReadTimeout)
	assert.Equal(t, props.WriteTimeout, result.WriteTimeout)
	assert.Equal(t, props.CallTimeout, result.CallTimeout)
	assert.Equal(t, props.PingInterval, result.PingInterval)
	assert.Equal(t, props.RetryOnConnectionFailure, result.RetryOnConnectionFailure)
}

func TestHttpClientProperties_Map_RetryOnConnectionFailure_False(t *testing.T) {
	props := &HttpClientProperties{
		MaxIdleConnections:       5,
		KeepAliveDuration:        time.Second,
		ConnectTimeout:           time.Second,
		ReadTimeout:              time.Second,
		WriteTimeout:             time.Second,
		CallTimeout:              time.Second,
		PingInterval:             time.Second,
		RetryOnConnectionFailure: false,
	}

	result, err := props.Map()

	require.NoError(t, err)
	assert.False(t, result.RetryOnConnectionFailure)
}
