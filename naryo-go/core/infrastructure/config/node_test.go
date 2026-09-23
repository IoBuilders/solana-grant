//go:build test

package config

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/common"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/node"
)

const validNodeID = "00000000-0000-0000-0000-000000000010"

// validWsPollNode returns a minimal valid NodeProperties (Solana / WS / POLL).
// Individual tests override specific fields to exercise error branches.
func validWsPollNode() NodeProperties {
	return NodeProperties{
		ID:   validNodeID,
		Name: "test-node",
		Type: node.TypeSolana.String(),
		Connection: NodeConnectionProperties{
			Type:     common.ConnectionTypeWs.String(),
			Endpoint: NodeEndpointProperties{URL: "ws://localhost:8900"},
			Retry: RetryConfigurationProperties{
				MaxRetries:   5,
				InitialDelay: time.Second,
				MaxDelay:     30 * time.Second,
				Multiplier:   2.0,
			},
		},
		Subscription: NodeSubscriptionProperties{
			Method:   node.BlockSubscriptionMethodPoll.String(),
			Interval: time.Second,
		},
	}
}

// --- error paths ---

func TestNodeProperties_Map_InvalidID(t *testing.T) {
	n := validWsPollNode()
	n.ID = "not-a-uuid"

	_, err := n.Map()

	assert.Error(t, err)
}

func TestNodeProperties_Map_UnsupportedSubscriptionMethod(t *testing.T) {
	n := validWsPollNode()
	n.Subscription.Method = "KAFKA"

	_, err := n.Map()

	assert.ErrorContains(t, err, "invalid subscription method: KAFKA")
}

func TestNodeProperties_Map_PollSubscription_ZeroInterval(t *testing.T) {
	n := validWsPollNode()
	n.Subscription.Interval = 0

	_, err := n.Map()

	assert.Error(t, err)
}

func TestNodeProperties_Map_InvalidRetryConfig(t *testing.T) {
	n := validWsPollNode()
	n.Connection.Retry.InitialDelay = 0 // must be > 0

	_, err := n.Map()

	assert.Error(t, err)
}

func TestNodeProperties_Map_EmptyConnectionEndpoint(t *testing.T) {
	n := validWsPollNode()
	n.Connection.Endpoint.URL = ""

	_, err := n.Map()

	assert.Error(t, err)
}

func TestNodeProperties_Map_UnsupportedConnectionType(t *testing.T) {
	n := validWsPollNode()
	n.Connection.Type = "TCP"

	_, err := n.Map()

	assert.ErrorContains(t, err, "invalid connection type: TCP")
}

func TestNodeProperties_Map_UnsupportedNodeType(t *testing.T) {
	n := validWsPollNode()
	n.Type = "EVM"

	_, err := n.Map()

	assert.ErrorContains(t, err, "unsupported node type: EVM")
}

func TestNodeProperties_Map_PubSubOverHTTP_Rejected(t *testing.T) {
	n := validWsPollNode()
	n.Connection.Type = common.ConnectionTypeHttp.String()
	n.Connection.Endpoint.URL = "http://localhost:8899"
	n.Subscription.Method = node.BlockSubscriptionMethodPubSub.String()

	_, err := n.Map()

	assert.Error(t, err)
}

// --- happy paths ---

func TestNodeProperties_Map_WsConnection_PollSubscription_Success(t *testing.T) {
	n := validWsPollNode()

	result, err := n.Map()

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, node.TypeSolana, result.Type)
	assert.Equal(t, node.Name("test-node"), result.Name)
	assert.Equal(t, common.ConnectionTypeWs, result.Connection.ConnectionType())
	assert.Equal(t, node.BlockSubscriptionMethodPoll, result.Subscription.MethodConfiguration.Method())
}

func TestNodeProperties_Map_WsConnection_PubSubSubscription_Success(t *testing.T) {
	n := validWsPollNode()
	n.Subscription.Method = node.BlockSubscriptionMethodPubSub.String()

	result, err := n.Map()

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, node.BlockSubscriptionMethodPubSub, result.Subscription.MethodConfiguration.Method())
}

func TestNodeProperties_Map_HttpConnection_PollSubscription_Success(t *testing.T) {
	n := validWsPollNode()
	n.Connection.Type = common.ConnectionTypeHttp.String()
	n.Connection.Endpoint.URL = "http://localhost:8899"

	result, err := n.Map()

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, common.ConnectionTypeHttp, result.Connection.ConnectionType())
	assert.Equal(t, node.BlockSubscriptionMethodPoll, result.Subscription.MethodConfiguration.Method())
}
