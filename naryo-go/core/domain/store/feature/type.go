package feature

// Type discriminates which concern a Configuration configures.
type Type string

const (
	TypeEvent       Type = "EVENT"
	TypeFilterSync  Type = "FILTER_SYNC"
	TypeLatestBlock Type = "LATEST_BLOCK"
)

func (t Type) IsValid() bool {
	switch t {
	case TypeEvent, TypeFilterSync, TypeLatestBlock:
		return true
	}
	return false
}

func (t Type) String() string {
	return string(t)
}
