package event

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/utils"
)

// DefaultListenerMaxConcurrency is the per-listener concurrency cap applied when
// no explicit limit is provided. It bounds how many goroutines may run the same
// listener at once, protecting finite resources such as DB connections.
const DefaultListenerMaxConcurrency = 50

// ListenerOption configures a ListenerDefinition at construction time.
type ListenerOption func(*ListenerDefinition)

// WithConcurrencyLimit sets the maximum number of goroutines that may execute
// the listener concurrently. Values <= 0 fall back to DefaultListenerMaxConcurrency.
func WithConcurrencyLimit(maxConcurrency int) ListenerOption {
	return func(l *ListenerDefinition) {
		l.semaphore = newListenerSemaphore(maxConcurrency)
	}
}

// WithExecutionTimeout sets a per-execution deadline. When non-zero, the context
// passed to Execute is cancelled after d, causing the listener to return early
// with context.DeadlineExceeded. Zero means no timeout.
func WithExecutionTimeout(d time.Duration) ListenerOption {
	return func(l *ListenerDefinition) {
		l.executionTimeout = d
	}
}

// ListenerDefinition defines a single listener that handles an event
type ListenerDefinition struct {
	Name    string
	Execute func(context.Context, Event) error
	// semaphore caps how many goroutines may execute this listener concurrently.
	semaphore chan struct{}
	// executionTimeout caps a single execution; zero means no timeout.
	executionTimeout time.Duration
}

// NewListenerDefinition creates a listener with optional configuration via
// ListenerOption values. Defaults: DefaultListenerMaxConcurrency concurrency,
// no execution timeout.
//
// Example:
//
//	NewListenerDefinition("my-listener", handler,
//	    WithConcurrencyLimit(5),
//	    WithExecutionTimeout(30*time.Second),
//	)
func NewListenerDefinition(name string, execute func(context.Context, Event) error, opts ...ListenerOption) ListenerDefinition {
	l := ListenerDefinition{
		Name:      name,
		Execute:   execute,
		semaphore: newListenerSemaphore(0),
	}
	for _, opt := range opts {
		opt(&l)
	}
	return l
}

// newListenerSemaphore builds the buffered channel that backs a listener's
// concurrency limit.
func newListenerSemaphore(maxConcurrency int) chan struct{} {
	if maxConcurrency <= 0 {
		maxConcurrency = DefaultListenerMaxConcurrency
	}
	return make(chan struct{}, maxConcurrency)
}

// ListenerRegistry maps event types to their listener definitions
type ListenerRegistry struct {
	listeners map[string][]ListenerDefinition
	types     map[string]reflect.Type
}

func NewListenerRegistry() *ListenerRegistry {
	return &ListenerRegistry{
		listeners: make(map[string][]ListenerDefinition),
		types:     make(map[string]reflect.Type),
	}
}

// Register adds listeners for a specific event type
func (r *ListenerRegistry) Register(eventType string, listeners ...ListenerDefinition) {
	for _, listener := range listeners {
		if !r.listenerExists(eventType, listener.Name) {
			r.listeners[eventType] = append(r.listeners[eventType], listener)
		}
	}
}

func (r *ListenerRegistry) listenerExists(eventType string, name string) bool {
	for _, l := range r.listeners[eventType] {
		if l.Name == name {
			return true
		}
	}
	return false
}

// RegisterType registers a Go type for a given event type name.
func (r *ListenerRegistry) RegisterType(eventType string, prototype interface{}) {
	if _, ok := r.types[eventType]; !ok {
		r.types[eventType] = utils.GetType(prototype)
	}
}

// Same as RegisterType but only with one parameter
func (r *ListenerRegistry) RegisterPrototype(prototype interface{}) {
	key := utils.GetType(prototype).String()
	if _, ok := r.types[key]; !ok {
		r.types[key] = utils.GetType(prototype)
	}
}

// CreateInstance creates a new instance of the registered event type.
func (r *ListenerRegistry) CreateInstance(eventType string) (interface{}, error) {
	t, ok := r.types[eventType]
	if !ok {
		// Try to see if it is in the listeners map, although we don't have the type there.
		return nil, fmt.Errorf("unknown event type: %s", eventType)
	}
	return reflect.New(t).Interface(), nil
}

// GetListeners returns all listeners registered for an event
func (r *ListenerRegistry) GetListeners(ev Event) []ListenerDefinition {
	eventType := utils.GetType(ev).String()
	return r.listeners[eventType]
}

func (r *ListenerRegistry) GetListenersByEventType(eventType string) []ListenerDefinition {
	return r.listeners[eventType]
}

// GetEventTypes returns all registered event types
func (r *ListenerRegistry) GetEventTypes() []string {
	types := make([]string, 0, len(r.listeners))
	for eventType := range r.listeners {
		types = append(types, eventType)
	}
	return types
}
