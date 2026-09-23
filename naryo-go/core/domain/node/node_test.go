//go:build test

package node

import (
	"testing"
	"time"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/common"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func newWsEndpoint(t *testing.T) *common.ConnectionEndpoint {
	t.Helper()
	endpoint, err := common.NewConnectionEndpointFromURL("wss://api.mainnet-beta.solana.com")
	assert.NoError(t, err)
	return endpoint
}

func newHTTPEndpoint(t *testing.T) *common.ConnectionEndpoint {
	t.Helper()
	endpoint, err := common.NewConnectionEndpointFromURL("https://api.mainnet-beta.solana.com")
	assert.NoError(t, err)
	return endpoint
}

func newValidNodeArgs(t *testing.T) (Name, common.Connection, BlockSubscriptionConfiguration) {
	t.Helper()
	name, err := NewName("solana-mainnet")
	assert.NoError(t, err)
	connection, err := common.NewWsConnection(newWsEndpoint(t), common.DefaultRetryConfiguration())
	assert.NoError(t, err)
	subscription, err := NewBlockSubscriptionConfiguration(NewPubSubBlockSubscriptionMethodConfiguration(), 0)
	assert.NoError(t, err)
	return name, connection, subscription
}

func newPollSubscription(t *testing.T) BlockSubscriptionConfiguration {
	t.Helper()
	poll, err := NewPollBlockSubscriptionMethodConfiguration(time.Second)
	assert.NoError(t, err)
	subscription, err := NewBlockSubscriptionConfiguration(poll, 0)
	assert.NoError(t, err)
	return subscription
}

func TestNode_NewSolanaNode(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		id := uuid.New()
		name, connection, subscription := newValidNodeArgs(t)

		n, err := NewSolanaNode(id, name, connection, subscription)
		assert.NoError(t, err)
		assert.Equal(t, id, n.ID)
		assert.Equal(t, name, n.Name)
		assert.Equal(t, TypeSolana, n.Type)
		assert.Equal(t, connection, n.Connection)
		assert.Equal(t, subscription, n.Subscription)
	})

	t.Run("ValidWithHttpConnectionAndPollMethod", func(t *testing.T) {
		name, _, _ := newValidNodeArgs(t)
		connection, err := common.NewHttpConnection(newHTTPEndpoint(t), common.DefaultRetryConfiguration())
		assert.NoError(t, err)
		subscription := newPollSubscription(t)

		n, err := NewSolanaNode(uuid.New(), name, connection, subscription)
		assert.NoError(t, err)
		assert.Equal(t, common.ConnectionTypeHttp, n.Connection.ConnectionType())
		assert.Equal(t, BlockSubscriptionMethodPoll, n.Subscription.MethodConfiguration.Method())
	})

	t.Run("ValidWithWsConnectionAndPollMethod", func(t *testing.T) {
		name, connection, _ := newValidNodeArgs(t)
		subscription := newPollSubscription(t)

		n, err := NewSolanaNode(uuid.New(), name, connection, subscription)
		assert.NoError(t, err)
		assert.Equal(t, common.ConnectionTypeWs, n.Connection.ConnectionType())
		assert.Equal(t, BlockSubscriptionMethodPoll, n.Subscription.MethodConfiguration.Method())
	})

	t.Run("PubSubMethodOverHttpConnection", func(t *testing.T) {
		name, _, subscription := newValidNodeArgs(t)
		connection, err := common.NewHttpConnection(newHTTPEndpoint(t), common.DefaultRetryConfiguration())
		assert.NoError(t, err)

		_, err = NewSolanaNode(uuid.New(), name, connection, subscription)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("NilID", func(t *testing.T) {
		name, connection, subscription := newValidNodeArgs(t)
		_, err := NewSolanaNode(uuid.Nil, name, connection, subscription)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("ZeroName", func(t *testing.T) {
		_, connection, subscription := newValidNodeArgs(t)
		_, err := NewSolanaNode(uuid.New(), "", connection, subscription)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("BlankName", func(t *testing.T) {
		_, connection, subscription := newValidNodeArgs(t)
		_, err := NewSolanaNode(uuid.New(), "   ", connection, subscription)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("NilConnection", func(t *testing.T) {
		name, _, subscription := newValidNodeArgs(t)
		_, err := NewSolanaNode(uuid.New(), name, nil, subscription)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("ZeroSubscription", func(t *testing.T) {
		name, connection, _ := newValidNodeArgs(t)
		_, err := NewSolanaNode(uuid.New(), name, connection, BlockSubscriptionConfiguration{})
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}
