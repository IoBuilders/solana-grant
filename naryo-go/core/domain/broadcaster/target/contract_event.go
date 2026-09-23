package target

// ContractEventTarget forwards every ingested contract event.
type ContractEventTarget struct {
	destinations []Destination
}

func NewContractEventTarget(destinations []Destination) (*ContractEventTarget, error) {
	t := &ContractEventTarget{destinations: destinations}
	if err := t.Validate(); err != nil {
		return nil, err
	}
	return t, nil
}

func (t ContractEventTarget) Type() Type {
	return TypeContractEvent
}

func (t ContractEventTarget) Destinations() []Destination {
	return t.destinations
}

func (t ContractEventTarget) Validate() error {
	return validateDestinations(t.destinations, "ContractEventTarget")
}

var _ Target = ContractEventTarget{}
