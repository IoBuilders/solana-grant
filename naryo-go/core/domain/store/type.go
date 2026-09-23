package store

// Type discriminates the backend an ActiveConfiguration persists to.
// TODO: future support: "MONGO", "HTTP".
type Type string

const (
	TypeGorm Type = "GORM"
)

func (t Type) IsValid() bool {
	switch t {
	case TypeGorm:
		return true
	}
	return false
}

func (t Type) String() string {
	return string(t)
}
