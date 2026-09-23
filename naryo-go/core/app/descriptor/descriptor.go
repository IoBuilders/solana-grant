package descriptor

type Descriptor[T any] interface {
	Map() (T, error)
}
