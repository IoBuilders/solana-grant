package node

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/common"
	"strings"

	"github.com/google/uuid"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

// Node is the aggregate root describing a blockchain node naryo ingests
// events from: its identity, the connection to reach it and the strategy used
// to subscribe to new blocks.
type Node struct {
	ID           uuid.UUID
	Name         Name
	Type         Type
	Connection   common.Connection
	Subscription BlockSubscriptionConfiguration
}

// NewSolanaNode builds a Solana Node. The remaining node types (EVM, HEDERA)
// will get their own constructors when supported.
func NewSolanaNode(id uuid.UUID, name Name, connection common.Connection, subscription BlockSubscriptionConfiguration) (*Node, error) {
	n := &Node{
		ID:           id,
		Name:         name,
		Type:         TypeSolana,
		Connection:   connection,
		Subscription: subscription,
	}
	if err := n.validate(); err != nil {
		return nil, err
	}
	return n, nil
}

func (n *Node) validate() error {
	if n.ID == uuid.Nil {
		return domainerrors.NewEmptyFieldError("ID", "Node")
	}
	if strings.TrimSpace(string(n.Name)) == "" {
		return domainerrors.NewEmptyFieldError("Name", "Node")
	}
	if !n.Type.IsValid() {
		return domainerrors.NewInvalidFieldError("Type", "Node", "unsupported node type "+n.Type.String())
	}
	if n.Connection == nil {
		return domainerrors.NewEmptyFieldError("Connection", "Node")
	}
	if err := n.Connection.Validate(); err != nil {
		return err
	}
	if n.Subscription.MethodConfiguration == nil {
		return domainerrors.NewEmptyFieldError("Subscription", "Node")
	}
	return n.validateSubscriptionMethod()
}

// validateSubscriptionMethod enforces the transport/method pairing: PUBSUB
// needs a streaming transport, so it cannot run over an HTTP connection.
// POLL works over both transports.
func (n *Node) validateSubscriptionMethod() error {
	method := n.Subscription.MethodConfiguration.Method()
	if n.Connection.ConnectionType() == common.ConnectionTypeHttp && method == BlockSubscriptionMethodPubSub {
		return domainerrors.NewInvalidFieldError(
			"Subscription", "Node",
			"PUBSUB subscription method is not supported over an HTTP connection",
		)
	}
	return nil
}
