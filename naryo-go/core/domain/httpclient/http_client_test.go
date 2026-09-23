//go:build test

package httpclient

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func validHttpClient() *HttpClient {
	return NewHttpClient(10, time.Second, time.Second, time.Second, time.Second, time.Second, time.Second, false)
}

func TestNewHttpClient_SetsAllFields(t *testing.T) {
	hc := NewHttpClient(5, time.Second, 2*time.Second, 3*time.Second, 4*time.Second, 5*time.Second, 6*time.Second, true)

	assert.Equal(t, 5, hc.MaxIdleConnections)
	assert.Equal(t, time.Second, hc.KeepAliveDuration)
	assert.Equal(t, 2*time.Second, hc.ConnectTimeout)
	assert.Equal(t, 3*time.Second, hc.ReadTimeout)
	assert.Equal(t, 4*time.Second, hc.WriteTimeout)
	assert.Equal(t, 5*time.Second, hc.CallTimeout)
	assert.Equal(t, 6*time.Second, hc.PingInterval)
	assert.True(t, hc.RetryOnConnectionFailure)
}

func TestHttpClient_Validate_Valid(t *testing.T) {
	err := validHttpClient().validate()

	require.NoError(t, err)
}

func TestHttpClient_Validate_ZeroMaxIdleConnections(t *testing.T) {
	hc := validHttpClient()
	hc.MaxIdleConnections = 0

	assert.ErrorIs(t, hc.validate(), domainerrors.ErrValidation)
}

func TestHttpClient_Validate_ZeroKeepAliveDuration(t *testing.T) {
	hc := validHttpClient()
	hc.KeepAliveDuration = 0

	assert.ErrorIs(t, hc.validate(), domainerrors.ErrValidation)
}

func TestHttpClient_Validate_ZeroConnectTimeout(t *testing.T) {
	hc := validHttpClient()
	hc.ConnectTimeout = 0

	assert.ErrorIs(t, hc.validate(), domainerrors.ErrValidation)
}

func TestHttpClient_Validate_ZeroReadTimeout(t *testing.T) {
	hc := validHttpClient()
	hc.ReadTimeout = 0

	assert.ErrorIs(t, hc.validate(), domainerrors.ErrValidation)
}

func TestHttpClient_Validate_ZeroWriteTimeout(t *testing.T) {
	hc := validHttpClient()
	hc.WriteTimeout = 0

	assert.ErrorIs(t, hc.validate(), domainerrors.ErrValidation)
}

func TestHttpClient_Validate_ZeroCallTimeout(t *testing.T) {
	hc := validHttpClient()
	hc.CallTimeout = 0

	assert.ErrorIs(t, hc.validate(), domainerrors.ErrValidation)
}

func TestHttpClient_Validate_ZeroPingInterval(t *testing.T) {
	hc := validHttpClient()
	hc.PingInterval = 0

	assert.ErrorIs(t, hc.validate(), domainerrors.ErrValidation)
}
