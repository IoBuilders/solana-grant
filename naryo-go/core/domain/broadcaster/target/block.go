package target

// BlockTarget forwards every ingested block.
type BlockTarget struct {
	destinations []Destination
}

func NewBlockTarget(destinations []Destination) (*BlockTarget, error) {
	t := &BlockTarget{destinations: destinations}
	if err := t.Validate(); err != nil {
		return nil, err
	}
	return t, nil
}

func (t BlockTarget) Type() Type {
	return TypeBlock
}

func (t BlockTarget) Destinations() []Destination {
	return t.destinations
}

func (t BlockTarget) Validate() error {
	return validateDestinations(t.destinations, "BlockTarget")
}

var _ Target = BlockTarget{}
