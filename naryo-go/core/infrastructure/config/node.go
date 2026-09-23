package config

import (
	"fmt"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/common"
	"time"

	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/node"
)

type NodeProperties struct {
	ID           string                     `mapstructure:"id"`
	Name         string                     `mapstructure:"name"`
	Type         string                     `mapstructure:"type"`
	Connection   NodeConnectionProperties   `mapstructure:"connection"`
	Subscription NodeSubscriptionProperties `mapstructure:"subscription"`
}

type NodeConnectionProperties struct {
	Type     string                       `mapstructure:"type"`
	Endpoint NodeEndpointProperties       `mapstructure:"endpoint"`
	Retry    RetryConfigurationProperties `mapstructure:"retry"`
}

type NodeEndpointProperties struct {
	URL string `mapstructure:"url"`
}

type RetryConfigurationProperties struct {
	MaxRetries   int           `mapstructure:"maxRetries"`
	InitialDelay time.Duration `mapstructure:"initialDelay"`
	MaxDelay     time.Duration `mapstructure:"maxDelay"`
	Multiplier   float64       `mapstructure:"multiplier"`
}

type NodeSubscriptionProperties struct {
	Method      string        `mapstructure:"method"`
	InitialSlot uint64        `mapstructure:"initialSlot"`
	Interval    time.Duration `mapstructure:"interval"`
}

func (n *NodeProperties) Map() (*node.Node, error) {
	id, err := uuid.Parse(n.ID)
	if err != nil {
		return nil, err
	}
	// Subcription Method
	var subscriptionMethod node.BlockSubscriptionMethodConfiguration
	switch n.Subscription.Method {
	case node.BlockSubscriptionMethodPoll.String():
		subscriptionMethod, err = node.NewPollBlockSubscriptionMethodConfiguration(n.Subscription.Interval)
		if err != nil {
			return nil, err
		}
	case node.BlockSubscriptionMethodPubSub.String():
		subscriptionMethod = node.NewPubSubBlockSubscriptionMethodConfiguration()
	default:
		return nil, fmt.Errorf("invalid subscription method: %s", n.Subscription.Method)
	}

	// Node connection
	retry, err := common.NewRetryConfiguration(n.Connection.Retry.MaxRetries, n.Connection.Retry.InitialDelay, n.Connection.Retry.MaxDelay, n.Connection.Retry.Multiplier)
	if err != nil {
		return nil, err
	}
	var connection common.Connection
	connectionEndpoint, err := common.NewConnectionEndpointFromURL(n.Connection.Endpoint.URL)
	if err != nil {
		return nil, err
	}
	var connectionErr error
	switch n.Connection.Type {
	case common.ConnectionTypeHttp.String():
		connection, connectionErr = common.NewHttpConnection(connectionEndpoint, retry)
	case common.ConnectionTypeWs.String():
		connection, connectionErr = common.NewWsConnection(connectionEndpoint, retry)
	default:
		return nil, fmt.Errorf("invalid connection type: %s", n.Connection.Type)
	}
	if connectionErr != nil {
		return nil, connectionErr
	}
	switch n.Type {
	case node.TypeSolana.String():
		subscription, err := node.NewBlockSubscriptionConfiguration(subscriptionMethod, n.Subscription.InitialSlot)
		if err != nil {
			return nil, err
		}
		return node.NewSolanaNode(id, node.Name(n.Name), connection, subscription)
	default:
		return nil, fmt.Errorf("unsupported node type: %s", n.Type)
	}
}
