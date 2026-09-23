// Package sourceprovider defines the SourceProvider port, an
// application-layer abstraction for retrieving raw configuration values
// of a given type T from an underlying source.
//
// A SourceProvider is agnostic to how its values are consumed; that
// concern belongs to a ConfigurationManager, which wraps a
// SourceProvider to add caching and any other orchestration logic.
//
// Concrete adapters (e.g. YAML file readers, remote config stores,
// in-memory fixtures) implement SourceProvider and are provided by the
// infrastructure layer, keeping this package free of any dependency on
// specific sourcing technology.
package sourceprovider
