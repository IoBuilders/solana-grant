package common

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

type Protocol string

const (
	ProtocolHTTP  Protocol = "HTTP"
	ProtocolHTTPS Protocol = "HTTPS"
	ProtocolWS    Protocol = "WS"
	ProtocolWSS   Protocol = "WSS"
)

func defaultPort(p Protocol) int {
	switch p {
	case ProtocolHTTP, ProtocolWS:
		return 80
	case ProtocolHTTPS, ProtocolWSS:
		return 443
	default:
		return 0
	}
}

// ConnectionEndpoint is the URL a Node connection dials. HTTP (http, https)
// and WebSocket (ws, wss) schemes are supported; each Connection kind
// additionally enforces the scheme family matching its transport.
type ConnectionEndpoint struct {
	Protocol Protocol
	Host     string
	Port     int
	Path     string
	Headers  map[string]string
}

func NewConnectionEndpointFromURL(rawURL string) (*ConnectionEndpoint, error) {
	return NewConnectionEndpoint(rawURL, nil)
}

// NewConnectionEndpoint parses a full URL, e.g. "https://example.com:8443/webhooks/events".
func NewConnectionEndpoint(rawURL string, headers map[string]string) (*ConnectionEndpoint, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, domainerrors.NewInvalidFieldError("URL", "ConnectionEndpoint", err.Error())
	}

	switch u.Scheme {
	case "http", "https", "ws", "wss":
	default:
		return nil, domainerrors.NewInvalidFieldError(
			"URL", "ConnectionEndpoint",
			fmt.Sprintf("unsupported scheme %q, must be one of http, https, ws, wss", u.Scheme),
		)
	}

	if u.Hostname() == "" {
		return nil, domainerrors.NewInvalidFieldError("Host", "ConnectionEndpoint", "host must not be empty")
	}

	protocol := Protocol(strings.ToUpper(u.Scheme))

	port := defaultPort(protocol)
	if u.Port() != "" {
		port, err = strconv.Atoi(u.Port())
		if err != nil {
			return nil, domainerrors.NewInvalidFieldError("Port", "ConnectionEndpoint", err.Error())
		}
	}

	path := strings.TrimPrefix(u.Path, "/")

	if headers == nil {
		headers = map[string]string{}
	}

	return &ConnectionEndpoint{
		Protocol: protocol,
		Host:     u.Hostname(),
		Port:     port,
		Path:     path,
		Headers:  headers,
	}, nil
}

// URL rebuilds the endpoint back into a URL string, omitting the port when it's the protocol default.
func (e ConnectionEndpoint) URL() string {
	result := fmt.Sprintf("%s://%s", strings.ToLower(string(e.Protocol)), e.Host)

	if e.Port != defaultPort(e.Protocol) {
		result += fmt.Sprintf(":%d", e.Port)
	}
	if e.Path != "" {
		result += CleanPath(e.Path)
	}
	return result
}

// Scheme returns the URL scheme, or an empty string if the URL is not
// parseable (only possible for endpoints built bypassing the constructor).
func (e ConnectionEndpoint) Scheme() string {
	parsed, err := url.Parse(e.URL())
	if err != nil {
		return ""
	}
	return parsed.Scheme
}

// CleanPath collapses repeated slashes and guarantees a single leading slash.
func CleanPath(path string) string {
	if path == "" {
		return "/"
	}
	collapsed := regexp.MustCompile(`/{2,}`).ReplaceAllString(path, "/")
	if !strings.HasPrefix(collapsed, "/") {
		collapsed = "/" + collapsed
	}
	return collapsed
}
