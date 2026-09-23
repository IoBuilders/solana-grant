package mapping

// Mapper converts a source value of type S into an output value of type O.
type Mapper[S, O any] interface {
	Map(source S) (O, error)
}
