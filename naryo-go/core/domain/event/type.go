package event

// Type discriminates the kind of event emitted while ingesting chain data.
type Type string

const (
	TypeBlock       Type = "BLOCK"
	TypeTransaction Type = "TRANSACTION"
	TypeContract    Type = "CONTRACT"
	TypeSlot        Type = "SLOT"
)

func (t Type) IsValid() bool {
	switch t {
	case TypeBlock, TypeTransaction, TypeContract, TypeSlot:
		return true
	}
	return false
}

func (t Type) String() string {
	return string(t)
}
