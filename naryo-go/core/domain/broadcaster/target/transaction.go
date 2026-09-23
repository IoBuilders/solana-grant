package target

// TransactionTarget forwards every ingested transaction.
type TransactionTarget struct {
	destinations []Destination
}

func NewTransactionTarget(destinations []Destination) (TransactionTarget, error) {
	t := TransactionTarget{destinations: destinations}
	if err := t.Validate(); err != nil {
		return TransactionTarget{}, err
	}
	return t, nil
}

func (t TransactionTarget) Type() Type {
	return TypeTransaction
}

func (t TransactionTarget) Destinations() []Destination {
	return t.destinations
}

func (t TransactionTarget) Validate() error {
	return validateDestinations(t.destinations, "TransactionTarget")
}

var _ Target = TransactionTarget{}
