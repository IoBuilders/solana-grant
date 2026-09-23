package event

// ContractEventStatus is the Solana confirmation level a ContractEvent was
// observed at, used by an EventFilter to constrain which events it matches.
type ContractEventStatus string

const (
	ContractEventStatusProcessed ContractEventStatus = "PROCESSED"
	ContractEventStatusConfirmed ContractEventStatus = "CONFIRMED"
	ContractEventStatusFinalized ContractEventStatus = "FINALIZED"
)

func (s ContractEventStatus) IsValid() bool {
	switch s {
	case ContractEventStatusProcessed, ContractEventStatusConfirmed, ContractEventStatusFinalized:
		return true
	}
	return false
}

func (s ContractEventStatus) String() string {
	return string(s)
}

func AllContractEventStatuses() []ContractEventStatus {
	return []ContractEventStatus{ContractEventStatusProcessed, ContractEventStatusConfirmed, ContractEventStatusFinalized}
}
