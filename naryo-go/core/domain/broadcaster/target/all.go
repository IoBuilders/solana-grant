package target

// AllTarget forwards every block, transaction and contract event.
type AllTarget struct {
	destinations []Destination
}

func NewAllTarget(destinations []Destination) (*AllTarget, error) {
	t := &AllTarget{destinations: destinations}
	if err := t.Validate(); err != nil {
		return nil, err
	}
	return t, nil
}

func (t AllTarget) Type() Type {
	return TypeAll
}

func (t AllTarget) Destinations() []Destination {
	return t.destinations
}

func (t AllTarget) Validate() error {
	return validateDestinations(t.destinations, "AllTarget")
}

var _ Target = AllTarget{}
