//go:build test

package configurationmanager

import (
	"context"
	"errors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/common"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/sourceprovider"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/node"
)

// --- stubs ---

type stubNodeSourceProvider struct {
	priority int
	items    []descriptor.Node
	err      error
}

func (s *stubNodeSourceProvider) Load(_ context.Context) ([]descriptor.Node, error) {
	return s.items, s.err
}

func (s *stubNodeSourceProvider) Priority() int { return s.priority }

type stubNodeDescriptor struct {
	result *node.Node
	err    error
}

func (s *stubNodeDescriptor) Map() (*node.Node, error) { return s.result, s.err }

func newNode(t *testing.T) *node.Node {
	t.Helper()
	endpoint, err := common.NewConnectionEndpointFromURL("ws://localhost:8900")
	require.NoError(t, err)
	conn, err := common.NewWsConnection(endpoint, common.DefaultRetryConfiguration())
	require.NoError(t, err)
	sub, err := node.NewBlockSubscriptionConfiguration(node.NewPubSubBlockSubscriptionMethodConfiguration(), 0)
	require.NoError(t, err)
	n, err := node.NewSolanaNode(uuid.New(), node.Name("test-node"), conn, sub)
	require.NoError(t, err)
	return n
}

// --- tests ---

func TestDefaultNodeConfigurationManager_Load_NoProviders(t *testing.T) {
	cm := NewDefaultNodeConfigurationManager(nil)

	result, err := cm.Load(context.Background())

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestDefaultNodeConfigurationManager_Load_SingleProvider(t *testing.T) {
	n1 := newNode(t)
	n2 := newNode(t)
	provider := &stubNodeSourceProvider{
		priority: 1,
		items: []descriptor.Node{
			&stubNodeDescriptor{result: n1},
			&stubNodeDescriptor{result: n2},
		},
	}
	cm := NewDefaultNodeConfigurationManager([]sourceprovider.NodeSourceProvider{provider})

	result, err := cm.Load(context.Background())

	require.NoError(t, err)
	assert.Equal(t, []*node.Node{n1, n2}, result)
}

func TestDefaultNodeConfigurationManager_Load_MultipleProviders_AggregatesAll(t *testing.T) {
	n1 := newNode(t)
	n2 := newNode(t)
	providerA := &stubNodeSourceProvider{
		priority: 1,
		items:    []descriptor.Node{&stubNodeDescriptor{result: n1}},
	}
	providerB := &stubNodeSourceProvider{
		priority: 2,
		items:    []descriptor.Node{&stubNodeDescriptor{result: n2}},
	}
	cm := NewDefaultNodeConfigurationManager([]sourceprovider.NodeSourceProvider{providerA, providerB})

	result, err := cm.Load(context.Background())

	require.NoError(t, err)
	assert.Equal(t, []*node.Node{n1, n2}, result)
}

func TestDefaultNodeConfigurationManager_Load_SortsByPriority(t *testing.T) {
	first := newNode(t)
	second := newNode(t)
	// Passed in reverse priority order — Load must still yield first before second.
	providerHigh := &stubNodeSourceProvider{
		priority: 2,
		items:    []descriptor.Node{&stubNodeDescriptor{result: second}},
	}
	providerLow := &stubNodeSourceProvider{
		priority: 1,
		items:    []descriptor.Node{&stubNodeDescriptor{result: first}},
	}
	cm := NewDefaultNodeConfigurationManager([]sourceprovider.NodeSourceProvider{providerHigh, providerLow})

	result, err := cm.Load(context.Background())

	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, first, result[0])
	assert.Equal(t, second, result[1])
}

func TestDefaultNodeConfigurationManager_Load_ProviderError(t *testing.T) {
	providerErr := errors.New("source unavailable")
	provider := &stubNodeSourceProvider{priority: 1, err: providerErr}
	cm := NewDefaultNodeConfigurationManager([]sourceprovider.NodeSourceProvider{provider})

	_, err := cm.Load(context.Background())

	assert.ErrorIs(t, err, providerErr)
}

func TestDefaultNodeConfigurationManager_Load_DescriptorMapError(t *testing.T) {
	mapErr := errors.New("invalid descriptor")
	provider := &stubNodeSourceProvider{
		priority: 1,
		items:    []descriptor.Node{&stubNodeDescriptor{err: mapErr}},
	}
	cm := NewDefaultNodeConfigurationManager([]sourceprovider.NodeSourceProvider{provider})

	_, err := cm.Load(context.Background())

	assert.ErrorIs(t, err, mapErr)
}
