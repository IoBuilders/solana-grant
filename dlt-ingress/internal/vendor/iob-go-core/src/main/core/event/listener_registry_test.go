package event

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/utils"
)

type TestEvent struct {
	BaseEvent
}

func TestCoreListenerRegistry_Register(t *testing.T) {
	registry := NewListenerRegistry()
	eventType := "TestEvent"
	listener := NewListenerDefinition("test-listener", func(ctx context.Context, ev Event) error { return nil })

	registry.Register(eventType, listener)

	assert.Len(t, registry.listeners[eventType], 1)
	assert.Equal(t, "test-listener", registry.listeners[eventType][0].Name)
}

func TestCoreListenerRegistry_Register_Duplicate(t *testing.T) {
	registry := NewListenerRegistry()
	eventType := "TestEvent"
	listener := NewListenerDefinition("test-listener", func(ctx context.Context, ev Event) error { return nil })

	registry.Register(eventType, listener)
	registry.Register(eventType, listener)

	assert.Len(t, registry.listeners[eventType], 1)
}

func TestCoreListenerRegistry_RegisterType_Duplicate(t *testing.T) {
	registry := NewListenerRegistry()
	eventType := "TestEvent"
	prototype := &TestEvent{}

	registry.RegisterType(eventType, prototype)
	registry.RegisterType(eventType, prototype)

	assert.Len(t, registry.types, 1)
}

func TestCoreListenerRegistry_RegisterType(t *testing.T) {
	registry := NewListenerRegistry()
	eventType := "TestEvent"
	prototype := &TestEvent{}

	registry.RegisterType(eventType, prototype)

	assert.Equal(t, reflect.TypeOf(TestEvent{}), registry.types[eventType])
}

func TestCoreListenerRegistry_CreateInstance(t *testing.T) {
	registry := NewListenerRegistry()
	eventType := "TestEvent"
	registry.RegisterType(eventType, &TestEvent{})

	instance, err := registry.CreateInstance(eventType)

	assert.NoError(t, err)
	assert.IsType(t, &TestEvent{}, instance)
}

func TestCoreListenerRegistry_CreateInstance_Unknown(t *testing.T) {
	registry := NewListenerRegistry()

	instance, err := registry.CreateInstance("Unknown")

	assert.Error(t, err)
	assert.Nil(t, instance)
	assert.Contains(t, err.Error(), "unknown event type")
}

func TestCoreListenerRegistry_GetListeners(t *testing.T) {
	registry := NewListenerRegistry()
	event := &TestEvent{}
	eventType := utils.GetType(event).String()
	listener := NewListenerDefinition("test-listener", func(ctx context.Context, ev Event) error { return nil })

	registry.Register(eventType, listener)

	listeners := registry.GetListeners(event)

	assert.Len(t, listeners, 1)
	assert.Equal(t, "test-listener", listeners[0].Name)
}

func TestCoreListenerRegistry_GetEventTypes(t *testing.T) {
	registry := NewListenerRegistry()
	registry.Register("Type1", NewListenerDefinition("L1", func(ctx context.Context, e Event) error { return nil }))
	registry.Register("Type2", NewListenerDefinition("L2", func(ctx context.Context, e Event) error { return nil }))

	types := registry.GetEventTypes()

	assert.Len(t, types, 2)
	assert.Contains(t, types, "Type1")
	assert.Contains(t, types, "Type2")
}
