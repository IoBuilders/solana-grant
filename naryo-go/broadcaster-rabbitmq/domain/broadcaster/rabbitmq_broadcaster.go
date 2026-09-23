package broadcaster

import (
	"fmt"
	"strings"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

const defaultVirtualHost = "/"

const defaultTLSAlgorithm = TLSAlgorithmTLS12

type RabbitMQBroadcaster struct {
	Host         string
	Port         int
	VirtualHost  string
	Username     string
	Password     string
	TLSEnabled   bool
	TLSAlgorithm TLSAlgorithm
}

func NewRabbitMQBroadcaster(host string, port int, virtualHost, username, password string, tlsEnabled bool, tlsAlgorithm TLSAlgorithm) (*RabbitMQBroadcaster, error) {
	if strings.TrimSpace(virtualHost) == "" {
		virtualHost = defaultVirtualHost
	}
	if tlsEnabled && strings.TrimSpace(tlsAlgorithm.String()) == "" {
		tlsAlgorithm = defaultTLSAlgorithm
	}

	b := RabbitMQBroadcaster{
		Host:         host,
		Port:         port,
		VirtualHost:  virtualHost,
		Username:     username,
		Password:     password,
		TLSEnabled:   tlsEnabled,
		TLSAlgorithm: tlsAlgorithm,
	}

	if err := b.validate(); err != nil {
		return nil, err
	}

	return &b, nil
}

func (b RabbitMQBroadcaster) validate() error {
	if strings.TrimSpace(b.Host) == "" {
		return domainerrors.NewEmptyFieldError("Host", "RabbitMQBroadcaster")
	}
	if b.Port < 1 || b.Port > 65535 {
		return domainerrors.NewInvalidFieldError("Port", "RabbitMQBroadcaster", "port must be between 1 and 65535")
	}
	if strings.TrimSpace(b.Username) == "" {
		return domainerrors.NewEmptyFieldError("Username", "RabbitMQBroadcaster")
	}
	if strings.TrimSpace(b.Password) == "" {
		return domainerrors.NewEmptyFieldError("Password", "RabbitMQBroadcaster")
	}
	if !b.TLSEnabled && strings.TrimSpace(b.TLSAlgorithm.String()) != "" {
		return domainerrors.NewInvalidFieldError(
			"TLSAlgorithm", "RabbitMQBroadcaster",
			"TLS algorithm must not be set while TLS is disabled",
		)
	}
	if b.TLSEnabled && !b.TLSAlgorithm.IsValid() {
		return domainerrors.NewInvalidFieldError(
			"TLSAlgorithm", "RabbitMQBroadcaster",
			fmt.Sprintf("unsupported TLS algorithm %q, must be one of %s, %s", b.TLSAlgorithm, TLSAlgorithmTLS12, TLSAlgorithmTLS13),
		)
	}
	return nil
}
