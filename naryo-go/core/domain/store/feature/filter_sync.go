package feature

type FilterSyncConfiguration struct {
	destination Destination
}

func NewFilterSyncConfiguration(destination Destination) (*FilterSyncConfiguration, error) {
	c := &FilterSyncConfiguration{destination: destination}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c FilterSyncConfiguration) Type() Type {
	return TypeFilterSync
}

func (c FilterSyncConfiguration) Destination() Destination {
	return c.destination
}

func (c FilterSyncConfiguration) Validate() error {
	return nil
}

var _ Configuration = (*FilterSyncConfiguration)(nil)
