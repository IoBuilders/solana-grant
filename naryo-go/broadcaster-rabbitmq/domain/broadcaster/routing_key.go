package broadcaster

import (
	"regexp"
	"strings"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

const maxRoutingKeyBytes = 255

var routingKeyPattern = regexp.MustCompile(`\A[A-Za-z0-9_-]+(?:\.[A-Za-z0-9_-]+)*\z`)

type RoutingKey string

func NewRoutingKey(value string) (RoutingKey, error) {
	if strings.TrimSpace(value) == "" {
		return "", domainerrors.NewEmptyFieldError("Value", "RoutingKey")
	}
	if len(value) > maxRoutingKeyBytes {
		return "", domainerrors.NewInvalidFieldError("Value", "RoutingKey", "routing key cannot be longer than 255 bytes")
	}
	if !routingKeyPattern.MatchString(value) {
		return "", domainerrors.NewInvalidFieldError(
			"Value", "RoutingKey",
			"routing key must be dot-separated words containing letters, digits, '_' or '-'",
		)
	}
	return RoutingKey(value), nil
}

func (k RoutingKey) String() string {
	return string(k)
}
