package feature

type Configuration interface {
	Type() Type
	Validate() error
}
