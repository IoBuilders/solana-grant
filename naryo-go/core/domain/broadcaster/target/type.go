package target

// Type discriminates which events a Broadcaster forwards.
type Type string

const (
	TypeBlock         Type = "BLOCK"
	TypeTransaction   Type = "TRANSACTION"
	TypeContractEvent Type = "CONTRACT_EVENT"
	TypeFilter        Type = "FILTER"
	TypeAll           Type = "ALL"
)

func (t Type) IsValid() bool {
	switch t {
	case TypeBlock, TypeTransaction, TypeContractEvent, TypeFilter, TypeAll:
		return true
	}
	return false
}

func (t Type) String() string {
	return string(t)
}
