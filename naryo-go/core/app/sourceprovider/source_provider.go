package sourceprovider

import "context"

// SourceProvider is a port defined by the application layer.
// Any adapter (file.yaml, HTTP, database, in-memory, etc.) that wants to
// supply configuration values of type T must implement this interface.
type SourceProvider[T any] interface {
	// Retrieves T from the underlying source.
	Load(ctx context.Context) (T, error)

	// Priority of the source provider
	Priority() int
}

// Source provider for collections
type CollectionSourceProvider[T any] interface {
	SourceProvider[[]T]
}
