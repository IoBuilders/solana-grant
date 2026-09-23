package filter

// Scope discriminates what a Filter matches against: a single program
// (CONTRACT) or every program ingested from the Node (GLOBAL).
type Scope string

const (
	ScopeContract Scope = "CONTRACT"
	ScopeGlobal   Scope = "GLOBAL"
)

func (s Scope) IsValid() bool {
	switch s {
	case ScopeContract, ScopeGlobal:
		return true
	}
	return false
}

func (s Scope) String() string {
	return string(s)
}
