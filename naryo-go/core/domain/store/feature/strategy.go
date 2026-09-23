package feature

// Strategy is how an event store decides what to persist.
type Strategy string

const (
	StrategyBlockBased Strategy = "BLOCK_BASED"
)

func (s Strategy) IsValid() bool {
	switch s {
	case StrategyBlockBased:
		return true
	}
	return false
}

func (s Strategy) String() string {
	return string(s)
}
