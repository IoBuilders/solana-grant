//go:build test

package configurationmanager

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-kafka/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-kafka/app/sourceprovider"
	"gitlab.com/iobuilders/projects/eng/naryo-go/broadcaster-kafka/domain/broadcaster"
)

// --- stubs ---

type stubKafkaBroadcasterSourceProvider struct {
	item descriptor.KafkaBroadcaster
	err  error
}

func (s *stubKafkaBroadcasterSourceProvider) Load(_ context.Context) (descriptor.KafkaBroadcaster, error) {
	return s.item, s.err
}

func (s *stubKafkaBroadcasterSourceProvider) Priority() int { return 1 }

type stubKafkaBroadcasterDescriptor struct {
	result *broadcaster.KafkaBroadcaster
	err    error
}

func (s *stubKafkaBroadcasterDescriptor) Map() (*broadcaster.KafkaBroadcaster, error) {
	return s.result, s.err
}

// --- tests ---

func TestDefaultKafkaBroadcasterConfigurationManager_Load_Success(t *testing.T) {
	expected, err := broadcaster.NewKafkaBroadcaster([]string{"broker1:9092"})
	require.NoError(t, err)
	provider := &stubKafkaBroadcasterSourceProvider{
		item: &stubKafkaBroadcasterDescriptor{result: expected},
	}
	cm := NewDefaultKafkaBroadcasterConfigurationManager([]sourceprovider.KafkaBroadcasterSourceProvider{provider})

	result, err := cm.Load(context.Background())

	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestDefaultKafkaBroadcasterConfigurationManager_Load_ProviderError(t *testing.T) {
	providerErr := errors.New("source unavailable")
	provider := &stubKafkaBroadcasterSourceProvider{err: providerErr}
	cm := NewDefaultKafkaBroadcasterConfigurationManager([]sourceprovider.KafkaBroadcasterSourceProvider{provider})

	_, err := cm.Load(context.Background())

	assert.ErrorIs(t, err, providerErr)
}

func TestDefaultKafkaBroadcasterConfigurationManager_Load_DescriptorMapError(t *testing.T) {
	mapErr := errors.New("invalid descriptor")
	provider := &stubKafkaBroadcasterSourceProvider{
		item: &stubKafkaBroadcasterDescriptor{err: mapErr},
	}
	cm := NewDefaultKafkaBroadcasterConfigurationManager([]sourceprovider.KafkaBroadcasterSourceProvider{provider})

	_, err := cm.Load(context.Background())

	assert.ErrorIs(t, err, mapErr)
}
