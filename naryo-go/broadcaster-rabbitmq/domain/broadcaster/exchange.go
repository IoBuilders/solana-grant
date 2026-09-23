package broadcaster

import (
	"strings"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

type Exchange string

func NewExchange(value string) (Exchange, error) {
	if strings.TrimSpace(value) == "" {
		return "", domainerrors.NewEmptyFieldError("Exchange", "RabbitMQConfiguration")
	}
	return Exchange(value), nil
}

func (e Exchange) String() string {
	return string(e)
}
