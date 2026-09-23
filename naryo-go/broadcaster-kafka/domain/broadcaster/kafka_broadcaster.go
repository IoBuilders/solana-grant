package broadcaster

import (
	"net"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

type KafkaBroadcaster struct {
	Brokers []string
}

func NewKafkaBroadcaster(brokers []string) (*KafkaBroadcaster, error) {
	p := KafkaBroadcaster{
		Brokers: brokers,
	}

	if err := p.validate(); err != nil {
		return nil, err
	}

	return &p, nil
}

func (p KafkaBroadcaster) validate() error {
	if len(p.Brokers) == 0 {
		return domainerrors.NewEmptyFieldError("Brokers", "KafkaBroadcaster")
	}

	for _, broker := range p.Brokers {
		if _, port, err := net.SplitHostPort(broker); err != nil || port == "" {
			return domainerrors.NewInvalidFieldError("Brokers", "KafkaBroadcaster", "broker must be a host:port address")
		}
	}

	return nil
}
