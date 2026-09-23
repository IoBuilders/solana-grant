// Package configurationmanager defines the ConfigurationManager port,
// an application-layer abstraction for loading configuration values of
// a given type T.
//
// A ConfigurationManager is agnostic to where or how the underlying
// values are sourced (file, remote store, in-memory, etc.); that concern
// belongs to a SourceProvider adapter, injected into concrete
// implementations such as DefaultConfigurationManager.
//
// DefaultConfigurationManager wraps a source_provider.SourceProvider[T]
// and caches the loaded values in memory after the first successful
// Load call, using a double-checked locking pattern to avoid redundant
// loads under concurrent access.
package configurationmanager
