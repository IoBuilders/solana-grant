//go:build test

package broadcasterproducer

import (
	"context"
	"errors"
	"testing"

	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	corebroadcaster "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster/target"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/infrastructure/event_mapper"

	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-kafka/domain/broadcaster"
)

func newTestBroadcaster(t *testing.T, destinations ...string) corebroadcaster.Broadcaster {
	t.Helper()
	dests := make([]target.Destination, len(destinations))
	for i, d := range destinations {
		dests[i] = target.Destination(d)
	}
	trg, err := target.NewBlockTarget(dests)
	require.NoError(t, err)
	b, err := corebroadcaster.NewBroadcaster(uuid.New(), trg, uuid.New())
	require.NoError(t, err)
	return *b
}

func newTestEvent(t *testing.T) event.SolanaTransactionEvent {
	t.Helper()
	e, err := event.NewSolanaTransactionEvent(uuid.New(), "5abc", 1, nil, nil, nil, nil)
	require.NoError(t, err)
	return e
}

func newTestProducer(mockProducer sarama.SyncProducer, mapper *eventmapper.EventToJsonMapperMock) *KafkaBroadcasterProducer {
	return &KafkaBroadcasterProducer{
		producer: mockProducer,
		mapper:   mapper,
	}
}

func TestNewKafkaBroadcasterProducer_NoBrokers_ReturnsError(t *testing.T) {
	_, err := NewKafkaBroadcasterProducer(&broadcaster.KafkaBroadcaster{}, new(eventmapper.EventToJsonMapperMock))

	assert.ErrorContains(t, err, "creating kafka producer")
}

func TestKafkaBroadcasterProducer_Produce_SendsMappedPayloadToTopic(t *testing.T) {
	b := newTestBroadcaster(t, "topic1")
	e := newTestEvent(t)
	mappedPayload := []byte(`{"foo":"bar"}`)

	var capturedTopic string
	var capturedValue []byte
	mockProducer := mocks.NewSyncProducer(t, nil)
	mockProducer.ExpectSendMessageWithMessageCheckerFunctionAndSucceed(func(msg *sarama.ProducerMessage) error {
		capturedTopic = msg.Topic
		var err error
		capturedValue, err = msg.Value.Encode()
		return err
	})
	defer func() { assert.NoError(t, mockProducer.Close()) }()

	mapper := new(eventmapper.EventToJsonMapperMock)
	mapper.On("Map", eventmapper.BlockchainEventSource{Event: e, Broadcaster: b}).Return(mappedPayload, nil).Once()

	p := newTestProducer(mockProducer, mapper)

	err := p.Produce(context.Background(), b, nil, e)

	require.NoError(t, err)
	assert.Equal(t, "topic1", capturedTopic)
	assert.Equal(t, mappedPayload, capturedValue)
	mapper.AssertExpectations(t)
}

func TestKafkaBroadcasterProducer_Produce_MultipleDestinations_SendsToEach(t *testing.T) {
	b := newTestBroadcaster(t, "topic1", "topic2")

	mockProducer := mocks.NewSyncProducer(t, nil)
	mockProducer.ExpectSendMessageAndSucceed()
	mockProducer.ExpectSendMessageAndSucceed()
	defer func() { assert.NoError(t, mockProducer.Close()) }()

	mapper := new(eventmapper.EventToJsonMapperMock)
	mapper.On("Map", mock.Anything).Return([]byte(`{}`), nil).Once()

	p := newTestProducer(mockProducer, mapper)

	err := p.Produce(context.Background(), b, nil, newTestEvent(t))

	assert.NoError(t, err)
	// Mapping happens once and the result is reused for every destination,
	// not re-mapped per destination.
	mapper.AssertExpectations(t)
}

func TestKafkaBroadcasterProducer_Produce_BrokerError_Propagates(t *testing.T) {
	b := newTestBroadcaster(t, "topic1")

	mockProducer := mocks.NewSyncProducer(t, nil)
	mockProducer.ExpectSendMessageAndFail(sarama.ErrOutOfBrokers)
	defer func() { assert.NoError(t, mockProducer.Close()) }()

	mapper := new(eventmapper.EventToJsonMapperMock)
	mapper.On("Map", mock.Anything).Return([]byte(`{}`), nil).Once()

	p := newTestProducer(mockProducer, mapper)

	err := p.Produce(context.Background(), b, nil, newTestEvent(t))

	assert.ErrorContains(t, err, "topic1")
	assert.ErrorIs(t, err, sarama.ErrOutOfBrokers)
	mapper.AssertExpectations(t)
}

func TestKafkaBroadcasterProducer_Produce_MappingError_DoesNotProduce(t *testing.T) {
	b := newTestBroadcaster(t, "topic1")
	mapErr := errors.New("dummyError")

	mockProducer := mocks.NewSyncProducer(t, nil)
	defer func() { assert.NoError(t, mockProducer.Close()) }()

	mapper := new(eventmapper.EventToJsonMapperMock)
	mapper.On("Map", mock.Anything).Return(nil, mapErr).Once()

	p := newTestProducer(mockProducer, mapper)

	err := p.Produce(context.Background(), b, nil, newTestEvent(t))

	assert.ErrorIs(t, err, mapErr)
	mapper.AssertExpectations(t)
}

func TestKafkaBroadcasterProducer_Produce_ContextCancelled_DoesNotProduce(t *testing.T) {
	b := newTestBroadcaster(t, "topic1")

	mockProducer := mocks.NewSyncProducer(t, nil)
	defer func() { assert.NoError(t, mockProducer.Close()) }()

	mapper := new(eventmapper.EventToJsonMapperMock)
	mapper.On("Map", mock.Anything).Return([]byte(`{}`), nil).Once()

	p := newTestProducer(mockProducer, mapper)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := p.Produce(ctx, b, nil, newTestEvent(t))

	assert.ErrorIs(t, err, context.Canceled)
	// Mapping isn't guarded by the context, only the actual broker send is.
	mapper.AssertExpectations(t)
}

func TestKafkaBroadcasterProducer_Supports(t *testing.T) {
	mapper := new(eventmapper.EventToJsonMapperMock)
	p := newTestProducer(mocks.NewSyncProducer(t, nil), mapper)

	assert.True(t, p.Supports(broadcaster.TypeKafka))
	assert.False(t, p.Supports(corebroadcaster.TypeHTTP))
	mapper.AssertNotCalled(t, "Map", mock.Anything)
}
