//go:build test

package httpbroadcasterproducer

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-http/domain/broadcaster"
	corebroadcaster "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster/target"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/httpclient"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/infrastructure/event_mapper"
)

func newTestEvent(t *testing.T) event.SolanaTransactionEvent {
	t.Helper()
	e, err := event.NewSolanaTransactionEvent(uuid.New(), "5abc", 1, nil, nil, nil, nil)
	require.NoError(t, err)
	return e
}

func newTestProducer(mapper *eventmapper.EventToJsonMapperMock) *HttpProducer {
	return NewHttpProducer(&httpclient.HttpClient{
		MaxIdleConnections: 1,
		KeepAliveDuration:  time.Second,
		ConnectTimeout:     time.Second,
		ReadTimeout:        time.Second,
	}, mapper)
}

func newTestBroadcaster(t *testing.T) corebroadcaster.Broadcaster {
	t.Helper()
	trg, err := target.NewBlockTarget([]target.Destination{"/webhook"})
	require.NoError(t, err)
	b, err := corebroadcaster.NewBroadcaster(uuid.New(), trg, uuid.New())
	require.NoError(t, err)
	return *b
}

func newTestHTTPConfiguration(t *testing.T, url string) *broadcaster.HTTPConfiguration {
	t.Helper()
	additionalProperties := map[string]interface{}{
		"endpoint": map[string]interface{}{"url": url},
		"retry": map[string]interface{}{
			"maxRetries":   0,
			"initialDelay": "1ms",
			"maxDelay":     "1ms",
			"multiplier":   1.0,
		},
	}
	generic, err := corebroadcaster.NewGenericConfiguration(uuid.New(), broadcaster.TypeHTTP, additionalProperties)
	require.NoError(t, err)
	cfg, err := broadcaster.NewHTTPConfiguration(generic)
	require.NoError(t, err)
	return cfg
}

// TestHttpProducer_Produce_PointerConfiguration reproduces how the rest of
// the codebase actually calls Produce: NewHTTPConfiguration returns a
// *broadcaster.HTTPConfiguration, and that pointer is what implements
// broadcaster.Configuration.
func TestHttpProducer_Produce_PointerConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	b := newTestBroadcaster(t)
	e := newTestEvent(t)
	mapper := new(eventmapper.EventToJsonMapperMock)
	mapper.On("Map", eventmapper.BlockchainEventSource{Event: e, Broadcaster: b}).Return([]byte(`{}`), nil).Once()

	p := newTestProducer(mapper)
	cfg := newTestHTTPConfiguration(t, server.URL)

	err := p.Produce(context.Background(), b, cfg, e)

	assert.NoError(t, err)
	mapper.AssertExpectations(t)
}

// TestHttpProducer_Produce_ValueConfiguration shows the type assertion
// rejects the value type now that it's the pointer that implements
// broadcaster.Configuration in practice.
func TestHttpProducer_Produce_ValueConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	b := newTestBroadcaster(t)
	mapper := new(eventmapper.EventToJsonMapperMock)

	p := newTestProducer(mapper)
	cfg := newTestHTTPConfiguration(t, server.URL)

	err := p.Produce(context.Background(), b, *cfg, newTestEvent(t))

	assert.ErrorContains(t, err, "unexpected configuration type")
	mapper.AssertNotCalled(t, "Map", mock.Anything)
}

// TestHttpProducer_Produce_SendsMappedPayload verifies the request body is
// exactly what the mapper returned, not a raw json.Marshal of the domain
// event.
func TestHttpProducer_Produce_SendsMappedPayload(t *testing.T) {
	var body []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	b := newTestBroadcaster(t)
	e := newTestEvent(t)
	mappedPayload := []byte(`{"foo":"bar"}`)
	mapper := new(eventmapper.EventToJsonMapperMock)
	mapper.On("Map", eventmapper.BlockchainEventSource{Event: e, Broadcaster: b}).Return(mappedPayload, nil).Once()

	p := newTestProducer(mapper)
	cfg := newTestHTTPConfiguration(t, server.URL)

	err := p.Produce(context.Background(), b, cfg, e)

	require.NoError(t, err)
	assert.Equal(t, mappedPayload, body)
	mapper.AssertExpectations(t)
}

func TestHttpProducer_Produce_MappingError_DoesNotSend(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	b := newTestBroadcaster(t)
	mapErr := errors.New("mapping failed")
	mapper := new(eventmapper.EventToJsonMapperMock)
	mapper.On("Map", mock.Anything).Return(nil, mapErr).Once()

	p := newTestProducer(mapper)
	cfg := newTestHTTPConfiguration(t, server.URL)

	err := p.Produce(context.Background(), b, cfg, newTestEvent(t))

	assert.ErrorIs(t, err, mapErr)
	assert.False(t, called)
	mapper.AssertExpectations(t)
}

func TestHttpProducer_Supports(t *testing.T) {
	mapper := new(eventmapper.EventToJsonMapperMock)
	p := newTestProducer(mapper)

	assert.True(t, p.Supports(broadcaster.TypeHTTP))
	assert.False(t, p.Supports("KAFKA"))
	mapper.AssertNotCalled(t, "Map", mock.Anything)
}
