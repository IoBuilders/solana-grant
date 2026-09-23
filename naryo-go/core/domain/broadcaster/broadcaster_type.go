package broadcaster

// Type discriminates the transport a Broadcaster uses to forward events.
// core only defines the transports it ships adapters for (HTTP); other
// modules may define additional Type values of their own.
type Type string

const (
	TypeHTTP Type = "HTTP"
)

func (t Type) String() string {
	return string(t)
}
