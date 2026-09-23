package common

import (
	"fmt"
	"strings"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

// ConnectionType discriminates the transport a Node connection uses.
type ConnectionType string

const (
	ConnectionTypeWs   ConnectionType = "WS"
	ConnectionTypeHttp ConnectionType = "HTTP"
)

func (t ConnectionType) IsValid() bool {
	switch t {
	case ConnectionTypeWs, ConnectionTypeHttp:
		return true
	}
	return false
}

func (t ConnectionType) String() string {
	return string(t)
}

// Connection describes how a Node is reached. The unexported validate
// method closes the interface to implementations in this package.
type Connection interface {
	ConnectionType() ConnectionType
	ConnectionEndpoint() *ConnectionEndpoint
	RetryConfiguration() *RetryConfiguration
	Validate() error
}

// WsConnection connects to a Node over WebSocket (ws, wss).
type WsConnection struct {
	Endpoint *ConnectionEndpoint
	Retry    *RetryConfiguration
}

func NewWsConnection(endpoint *ConnectionEndpoint, retry *RetryConfiguration) (*WsConnection, error) {
	c := &WsConnection{Endpoint: endpoint, Retry: retry}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *WsConnection) ConnectionType() ConnectionType {
	return ConnectionTypeWs
}

func (c *WsConnection) ConnectionEndpoint() *ConnectionEndpoint {
	return c.Endpoint
}

func (c *WsConnection) RetryConfiguration() *RetryConfiguration {
	return c.Retry
}

func (c *WsConnection) Validate() error {
	if c.Endpoint.URL() == "" {
		return domainerrors.NewEmptyFieldError("Endpoint", "WsConnection")
	}
	switch scheme := c.Endpoint.Scheme(); scheme {
	case "ws", "wss":
	default:
		return domainerrors.NewInvalidFieldError(
			"Endpoint", "WsConnection",
			fmt.Sprintf("invalid scheme %q for a WS connection, must be ws or wss", scheme),
		)
	}
	return c.Retry.Validate()
}

// HttpConnection connects to a Node over HTTP (http, https).
type HttpConnection struct {
	Endpoint *ConnectionEndpoint
	Retry    *RetryConfiguration
}

func NewHttpConnection(endpoint *ConnectionEndpoint, retry *RetryConfiguration) (*HttpConnection, error) {
	c := &HttpConnection{Endpoint: endpoint, Retry: retry}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *HttpConnection) ConnectionType() ConnectionType {
	return ConnectionTypeHttp
}

func (c *HttpConnection) ConnectionEndpoint() *ConnectionEndpoint {
	return c.Endpoint
}

func (c *HttpConnection) RetryConfiguration() *RetryConfiguration {
	return c.Retry
}

func (c *HttpConnection) Validate() error {
	if c.Endpoint == nil || strings.TrimSpace(c.Endpoint.URL()) == "" {
		return domainerrors.NewEmptyFieldError("Endpoint", "HttpConnection")
	}
	switch scheme := c.Endpoint.Scheme(); scheme {
	case "http", "https":
	default:
		return domainerrors.NewInvalidFieldError(
			"Endpoint", "HttpConnection",
			fmt.Sprintf("invalid scheme %q for an HTTP connection, must be http or https", scheme),
		)
	}
	return c.Retry.Validate()
}

var (
	_ Connection = (*WsConnection)(nil)
	_ Connection = (*HttpConnection)(nil)
)
