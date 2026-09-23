//go:build test

package broadcasterproducer

import (
	"context"
	"crypto/tls"
	"errors"
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	corebroadcaster "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster/target"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/infrastructure/event_mapper"

	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-rabbitmq/domain/broadcaster"
)

// --- doubles ---

type publishCall struct {
	ctx        context.Context
	exchange   string
	routingKey string
	mandatory  bool
	immediate  bool
	msg        amqp091.Publishing
}

// fakePublisher records every publish and optionally fails for named routing keys.
type fakePublisher struct {
	calls      []publishCall
	failFor    map[string]error
	failAlways error
}

func (f *fakePublisher) PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp091.Publishing) error {
	f.calls = append(f.calls, publishCall{ctx, exchange, key, mandatory, immediate, msg})
	if f.failAlways != nil {
		return f.failAlways
	}
	return f.failFor[key]
}

// --- helpers ---

func toDestinations(destinations []string) []target.Destination {
	dests := make([]target.Destination, len(destinations))
	for i, d := range destinations {
		dests[i] = target.Destination(d)
	}
	return dests
}

func newTestBroadcaster(t *testing.T, destinations ...string) corebroadcaster.Broadcaster {
	t.Helper()
	trg, err := target.NewAllTarget(toDestinations(destinations))
	require.NoError(t, err)
	b, err := corebroadcaster.NewBroadcaster(uuid.New(), trg, uuid.New())
	require.NoError(t, err)
	return *b
}

func newTestFilterBroadcaster(t *testing.T, destinations ...string) corebroadcaster.Broadcaster {
	t.Helper()
	trg, err := target.NewFilterTarget(toDestinations(destinations), uuid.New())
	require.NoError(t, err)
	b, err := corebroadcaster.NewBroadcaster(uuid.New(), trg, uuid.New())
	require.NoError(t, err)
	return *b
}

func newTestConfiguration(t *testing.T, exchange string) *broadcaster.RabbitMQConfiguration {
	t.Helper()
	generic, err := corebroadcaster.NewGenericConfiguration(uuid.New(), broadcaster.TypeRabbitMQ, map[string]interface{}{
		"destination": map[string]interface{}{"exchange": exchange},
	})
	require.NoError(t, err)
	c, err := broadcaster.NewRabbitMQConfiguration(generic)
	require.NoError(t, err)
	return c
}

func newTestBlockEvent(t *testing.T) event.SolanaBlockEvent {
	t.Helper()
	e, err := event.NewSolanaBlockEvent(uuid.New(), 42, "blockhash", nil, nil)
	require.NoError(t, err)
	return e
}

func newTestContractEvent(t *testing.T) event.SolanaContractEvent {
	t.Helper()
	e, err := event.NewSolanaContractEvent(uuid.New(), "Prog111", "5abc", 1, nil, "MintDeployed", event.ContractEventStatusConfirmed)
	require.NoError(t, err)
	return e
}

func newTestEvent(t *testing.T) event.SolanaTransactionEvent {
	t.Helper()
	e, err := event.NewSolanaTransactionEvent(uuid.New(), "5abc", 1, nil, nil, nil, nil)
	require.NoError(t, err)
	return e
}

func newTestProducer(channel publisher, mapper *eventmapper.EventToJsonMapperMock) *RabbitMQBroadcasterProducer {
	return &RabbitMQBroadcasterProducer{channel: channel, mapper: mapper}
}

// --- Produce ---

func TestRabbitMQBroadcasterProducer_Produce_PublishesToConfigurationExchangeWithRoutingKey(t *testing.T) {
	b := newTestBroadcaster(t, "naryo.transactions")
	e := newTestEvent(t)
	payload := []byte(`{"foo":"bar"}`)

	mapper := new(eventmapper.EventToJsonMapperMock)
	mapper.On("Map", eventmapper.BlockchainEventSource{Event: e, Broadcaster: b}).Return(payload, nil).Once()
	channel := &fakePublisher{}

	err := newTestProducer(channel, mapper).Produce(context.Background(), b, newTestConfiguration(t, "naryo-events"), e)

	require.NoError(t, err)
	require.Len(t, channel.calls, 1)
	call := channel.calls[0]
	assert.Equal(t, "naryo-events", call.exchange, "the exchange comes from the configuration")
	assert.Equal(t, "naryo.transactions.5abc", call.routingKey, "the routing key is <destination>.<signature>")
	assert.False(t, call.mandatory)
	assert.False(t, call.immediate)
	assert.Equal(t, payload, call.msg.Body)
	assert.Equal(t, "application/json", call.msg.ContentType)
	mapper.AssertExpectations(t)
}

func TestRabbitMQBroadcasterProducer_Produce_PublishesOncePerDestination(t *testing.T) {
	b := newTestBroadcaster(t, "dest-a", "dest-b", "dest-c")
	e := newTestEvent(t)

	mapper := new(eventmapper.EventToJsonMapperMock)
	mapper.On("Map", eventmapper.BlockchainEventSource{Event: e, Broadcaster: b}).Return([]byte(`{}`), nil).Once()
	channel := &fakePublisher{}

	err := newTestProducer(channel, mapper).Produce(context.Background(), b, newTestConfiguration(t, "naryo-events"), e)

	require.NoError(t, err)
	require.Len(t, channel.calls, 3)
	for i, key := range []string{"dest-a.5abc", "dest-b.5abc", "dest-c.5abc"} {
		assert.Equal(t, "naryo-events", channel.calls[i].exchange)
		assert.Equal(t, key, channel.calls[i].routingKey)
	}
	mapper.AssertNumberOfCalls(t, "Map", 1)
}

func TestRabbitMQBroadcasterProducer_Produce_RoutingKeyPerEventType(t *testing.T) {
	tests := []struct {
		name        string
		broadcaster corebroadcaster.Broadcaster
		event       event.Event
		expected    string
	}{
		{"block uses the slot", newTestBroadcaster(t, "blocks"), newTestBlockEvent(t), "blocks.42"},
		{"transaction uses the signature", newTestBroadcaster(t, "txs"), newTestEvent(t), "txs.5abc"},
		{"contract event uses name and program", newTestBroadcaster(t, "contracts"), newTestContractEvent(t), "contracts.MintDeployed.Prog111"},
		{"contract event via filter target uses the program only", newTestFilterBroadcaster(t, "filtered"), newTestContractEvent(t), "filtered.Prog111"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapper := new(eventmapper.EventToJsonMapperMock)
			mapper.On("Map", mock.Anything).Return([]byte(`{}`), nil).Once()
			channel := &fakePublisher{}

			err := newTestProducer(channel, mapper).Produce(context.Background(), tt.broadcaster, newTestConfiguration(t, "naryo-events"), tt.event)

			require.NoError(t, err)
			require.Len(t, channel.calls, 1)
			assert.Equal(t, tt.expected, channel.calls[0].routingKey)
		})
	}
}

func TestRabbitMQBroadcasterProducer_Produce_InvalidRoutingKeyStillPublishesTheRest(t *testing.T) {
	b := newTestBroadcaster(t, "/events", "naryo.events")
	e := newTestEvent(t)

	mapper := new(eventmapper.EventToJsonMapperMock)
	mapper.On("Map", eventmapper.BlockchainEventSource{Event: e, Broadcaster: b}).Return([]byte(`{}`), nil).Once()
	channel := &fakePublisher{}

	err := newTestProducer(channel, mapper).Produce(context.Background(), b, newTestConfiguration(t, "naryo-events"), e)

	assert.ErrorIs(t, err, domainerrors.ErrValidation)
	assert.ErrorContains(t, err, "destination /events")
	require.Len(t, channel.calls, 1, "only the valid destination is published")
	assert.Equal(t, "naryo.events.5abc", channel.calls[0].routingKey)
}

func TestRabbitMQBroadcasterProducer_Produce_UnsupportedConfiguration_PublishesNothing(t *testing.T) {
	b := newTestBroadcaster(t, "naryo.events")
	generic, err := corebroadcaster.NewGenericConfiguration(uuid.New(), broadcaster.TypeRabbitMQ, nil)
	require.NoError(t, err)
	mapper := new(eventmapper.EventToJsonMapperMock)
	channel := &fakePublisher{}

	err = newTestProducer(channel, mapper).Produce(context.Background(), b, generic, newTestEvent(t))

	assert.ErrorContains(t, err, "unsupported configuration")
	assert.Empty(t, channel.calls)
	mapper.AssertNotCalled(t, "Map", mock.Anything)
}

func TestRabbitMQBroadcasterProducer_Produce_MapperError_PublishesNothing(t *testing.T) {
	b := newTestBroadcaster(t, "naryo.events")
	e := newTestEvent(t)
	mapErr := errors.New("cannot map event")

	mapper := new(eventmapper.EventToJsonMapperMock)
	mapper.On("Map", eventmapper.BlockchainEventSource{Event: e, Broadcaster: b}).Return(nil, mapErr).Once()
	channel := &fakePublisher{}

	err := newTestProducer(channel, mapper).Produce(context.Background(), b, newTestConfiguration(t, "naryo-events"), e)

	assert.ErrorIs(t, err, mapErr)
	assert.Empty(t, channel.calls, "no publish should be attempted when mapping fails")
}

func TestRabbitMQBroadcasterProducer_Produce_OneFailingDestinationStillTriesTheRest(t *testing.T) {
	b := newTestBroadcaster(t, "dest-a", "dest-b", "dest-c")
	e := newTestEvent(t)
	publishErr := errors.New("channel closed")

	mapper := new(eventmapper.EventToJsonMapperMock)
	mapper.On("Map", eventmapper.BlockchainEventSource{Event: e, Broadcaster: b}).Return([]byte(`{}`), nil).Once()
	channel := &fakePublisher{failFor: map[string]error{"dest-b.5abc": publishErr}}

	err := newTestProducer(channel, mapper).Produce(context.Background(), b, newTestConfiguration(t, "naryo-events"), e)

	require.Error(t, err)
	assert.ErrorIs(t, err, publishErr)
	assert.ErrorContains(t, err, "exchange naryo-events, routing key dest-b.5abc")
	assert.Len(t, channel.calls, 3, "a failing destination must not abort the others")
}

func TestRabbitMQBroadcasterProducer_Produce_JoinsEveryFailure(t *testing.T) {
	b := newTestBroadcaster(t, "dest-a", "dest-b")
	e := newTestEvent(t)
	publishErr := errors.New("broker unreachable")

	mapper := new(eventmapper.EventToJsonMapperMock)
	mapper.On("Map", eventmapper.BlockchainEventSource{Event: e, Broadcaster: b}).Return([]byte(`{}`), nil).Once()
	channel := &fakePublisher{failAlways: publishErr}

	err := newTestProducer(channel, mapper).Produce(context.Background(), b, newTestConfiguration(t, "naryo-events"), e)

	require.Error(t, err)
	assert.ErrorContains(t, err, "routing key dest-a.5abc")
	assert.ErrorContains(t, err, "routing key dest-b.5abc")
}

func TestRabbitMQBroadcasterProducer_Produce_PassesContextThrough(t *testing.T) {
	b := newTestBroadcaster(t, "naryo.events")
	e := newTestEvent(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mapper := new(eventmapper.EventToJsonMapperMock)
	mapper.On("Map", eventmapper.BlockchainEventSource{Event: e, Broadcaster: b}).Return([]byte(`{}`), nil).Once()
	channel := &fakePublisher{}

	_ = newTestProducer(channel, mapper).Produce(ctx, b, newTestConfiguration(t, "naryo-events"), e)

	require.Len(t, channel.calls, 1)
	assert.ErrorIs(t, channel.calls[0].ctx.Err(), context.Canceled,
		"the caller's context must reach PublishWithContext, which itself short-circuits when cancelled")
}

// --- Supports ---

func TestRabbitMQBroadcasterProducer_Supports(t *testing.T) {
	p := newTestProducer(&fakePublisher{}, nil)

	assert.True(t, p.Supports(broadcaster.TypeRabbitMQ))
	assert.False(t, p.Supports(corebroadcaster.TypeHTTP))
}

// --- constructor ---

func TestNewRabbitMQBroadcasterProducer_UnreachableBroker_ReturnsError(t *testing.T) {
	// Port 1 on localhost refuses connections, so Dial fails without a broker.
	b, err := broadcaster.NewRabbitMQBroadcaster("127.0.0.1", 1, "/", "guest", "guest", false, "")
	require.NoError(t, err)

	_, err = NewRabbitMQBroadcasterProducer(b, new(eventmapper.EventToJsonMapperMock))

	assert.ErrorContains(t, err, "connecting to rabbitmq")
}

// --- URI building ---

func TestAmqpURI(t *testing.T) {
	tests := []struct {
		name     string
		vhost    string
		scheme   string
		expected string
	}{
		{"default vhost yields an empty path", "/", "amqp", "amqp://guest:guest@localhost:5672"},
		{"named vhost becomes a path segment", "/naryo", "amqp", "amqp://guest:guest@localhost:5672/naryo"},
		{"vhost without a leading slash", "naryo", "amqp", "amqp://guest:guest@localhost:5672/naryo"},
		{"tls uses the amqps scheme", "/", "amqps", "amqps://guest:guest@localhost:5672"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &broadcaster.RabbitMQBroadcaster{
				Host: "localhost", Port: 5672, VirtualHost: tt.vhost,
				Username: "guest", Password: "guest",
			}

			assert.Equal(t, tt.expected, amqpURI(b, tt.scheme))
		})
	}
}

func TestAmqpURI_EscapesCredentials(t *testing.T) {
	b := &broadcaster.RabbitMQBroadcaster{
		Host: "localhost", Port: 5672, VirtualHost: "/",
		Username: "user@naryo", Password: "p@ss:w/rd",
	}

	uri := amqpURI(b, "amqp")

	assert.Contains(t, uri, "user%40naryo", "the username must be percent-escaped")
	assert.NotContains(t, uri, "p@ss:w/rd", "the raw password must not leak into the URI")
	assert.Equal(t, "localhost:5672", mustParseHost(t, uri))
}

func mustParseHost(t *testing.T, uri string) string {
	t.Helper()
	u, err := url.Parse(uri)
	require.NoError(t, err)
	return u.Host
}

// --- TLS mapping ---

func TestTLSMinVersion(t *testing.T) {
	v12, err := tlsMinVersion(broadcaster.TLSAlgorithmTLS12)
	require.NoError(t, err)
	assert.Equal(t, uint16(tls.VersionTLS12), v12)

	v13, err := tlsMinVersion(broadcaster.TLSAlgorithmTLS13)
	require.NoError(t, err)
	assert.Equal(t, uint16(tls.VersionTLS13), v13)

	_, err = tlsMinVersion(broadcaster.TLSAlgorithm("TLSv1.1"))
	assert.ErrorContains(t, err, `unsupported TLS algorithm "TLSv1.1"`)
}
