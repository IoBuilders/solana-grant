package node

// Type discriminates the kind of blockchain a Node points to.
//
// Reserved for future support: "EVM", "HEDERA".
type Type string

const (
	TypeSolana Type = "SOLANA"
)

func (t Type) IsValid() bool {
	switch t {
	case TypeSolana:
		return true
	}
	return false
}

func (t Type) String() string {
	return string(t)
}
