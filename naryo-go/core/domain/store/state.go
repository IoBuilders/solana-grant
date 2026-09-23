package store

// State tells whether a Node's data is persisted at all.
type State string

const (
	StateActive   State = "ACTIVE"
	StateInactive State = "INACTIVE"
)

func (s State) IsValid() bool {
	switch s {
	case StateActive, StateInactive:
		return true
	}
	return false
}

func (s State) String() string {
	return string(s)
}
