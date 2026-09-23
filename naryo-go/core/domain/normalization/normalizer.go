package normalization

// Normalizer canonicalizes a value of type T.
type Normalizer[T any] interface {
	Normalize(in T) (T, error)
}
