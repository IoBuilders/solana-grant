package feature

type LatestBlockConfiguration struct {
	destination Destination
}

func NewLatestBlockConfiguration(destination Destination) (*LatestBlockConfiguration, error) {
	c := &LatestBlockConfiguration{destination: destination}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c LatestBlockConfiguration) Type() Type {
	return TypeLatestBlock
}

func (c LatestBlockConfiguration) Destination() Destination {
	return c.destination
}

func (c LatestBlockConfiguration) Validate() error {
	return nil
}

var _ Configuration = (*LatestBlockConfiguration)(nil)
