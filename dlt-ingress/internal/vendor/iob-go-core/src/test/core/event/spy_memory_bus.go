package testevent

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/utils"
)

const (
	spyEventPollInterval = 100 * time.Millisecond
	spyEventPreviewCount = 3
)

type SpyEventBus struct {
	bus    event.Bus
	mu     sync.RWMutex
	events []event.Event
}

func NewSpyEventBus(bus event.Bus) *SpyEventBus {
	return &SpyEventBus{
		bus:    bus,
		events: []event.Event{},
		mu:     sync.RWMutex{},
	}
}

func (s *SpyEventBus) Publish(ctx context.Context, e event.Event) error {
	s.mu.Lock()
	s.events = append(s.events, e)
	s.mu.Unlock()

	logger.Info(fmt.Sprintf("[SPY-BUS] Event %s captured", reflect.TypeOf(e).String()))
	return s.bus.Publish(ctx, e)
}

func (s *SpyEventBus) CleanEvents() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = []event.Event{}
	logger.Info("[SPY-BUS] Event list cleared")

}

func (s *SpyEventBus) Events() []event.Event {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]event.Event{}, s.events...)
}

func (s *SpyEventBus) EventsOfType(typeName string) []event.Event {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []event.Event
	for _, e := range s.events {
		if utils.GetType(e).String() == typeName {
			result = append(result, e)
		}
	}
	return result
}

// Deprecated: avoid to use this method. Use ExpectEventProcess, ExpectEventDLT or ExpectEventByField instead
func ExpectEvent[T event.Event](t *testing.T, bus *SpyEventBus, timeout time.Duration) T {
	eventName := utils.GetType((*T)(nil)).String()
	if found, ok := findCapturedEvent[T](bus, eventName); ok {
		return found
	}

	foundChan, stopChan := asyncWaitEvent[T](bus, eventName)
	defer close(stopChan)

	select {
	case v := <-foundChan:
		return v
	case <-time.After(timeout):
		reportTimeoutEvents(t, bus, eventName, timeout)
		var zero T
		return zero
	}
}

func ExpectEventByField[T event.Event](t *testing.T, bus *SpyEventBus, fieldName string, expected any, timeout time.Duration) T {
	eventName := utils.GetType((*T)(nil)).String()
	if found, ok := findCapturedEventByField[T](bus, eventName, fieldName, expected); ok {
		return found
	}

	foundChan, stopChan := asyncWaitEventByField[T](bus, eventName, fieldName, expected)
	defer close(stopChan)

	select {
	case v := <-foundChan:
		return v
	case <-time.After(timeout):
		reportTimeoutEventsByField(t, bus, eventName, fieldName, expected, timeout)
		var zero T
		return zero
	}
}

func ExpectEventProcess[T event.Event](t *testing.T, bus *SpyEventBus, processId uuid.UUID, timeout time.Duration) T {
	return ExpectEventByField[T](t, bus, "ProcessId", processId, timeout)
}

func ExpectEventDLT[T event.Event](t *testing.T, bus *SpyEventBus, transactionHash string, timeout time.Duration) T {
	return ExpectEventByField[T](t, bus, "TransactionHash", transactionHash, timeout)
}

func findCapturedEvent[T any](bus *SpyEventBus, eventName string) (T, bool) {
	captured := bus.EventsOfType(eventName)
	for _, e := range captured {
		if typed, ok := e.(T); ok {
			return typed, true
		}
	}
	var zero T
	return zero, false
}

func findCapturedEventByField[T any](bus *SpyEventBus, eventName, fieldName string, expected any) (T, bool) {
	captured := bus.EventsOfType(eventName)
	for _, e := range captured {
		if typed, ok := e.(T); ok && hasFieldValue(typed, fieldName, expected) {
			return typed, true
		}
	}
	var zero T
	return zero, false
}

func asyncWaitEvent[T any](bus *SpyEventBus, eventName string) (chan T, chan struct{}) {
	found := make(chan T, 1)
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(spyEventPollInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				captured := bus.EventsOfType(eventName)
				for _, e := range captured {
					if typed, ok := e.(T); ok {
						select {
						case found <- typed:
						default:
						}
						return
					}
				}
			case <-stop:
				return
			}
		}
	}()
	return found, stop
}

func asyncWaitEventProcess[T any](bus *SpyEventBus, eventName string, processId uuid.UUID) (chan T, chan struct{}) {
	return asyncWaitEventByField[T](bus, eventName, "ProcessId", processId)
}

func asyncWaitEventDLT[T any](bus *SpyEventBus, eventName, transactionHash string) (chan T, chan struct{}) {
	return asyncWaitEventByField[T](bus, eventName, "TransactionHash", transactionHash)
}

func asyncWaitEventByField[T any](bus *SpyEventBus, eventName, fieldName string, expected any) (chan T, chan struct{}) {
	found := make(chan T, 1)
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(spyEventPollInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				captured := bus.EventsOfType(eventName)
				for _, e := range captured {
					if typed, ok := e.(T); ok && hasFieldValue(typed, fieldName, expected) {
						select {
						case found <- typed:
						default:
						}
						return
					}
				}
			case <-stop:
				return
			}
		}
	}()
	return found, stop
}

func hasFieldValue(target any, fieldName string, expected any) bool {
	fieldValue, ok := findStructFieldValue(target, fieldName)
	if !ok {
		return false
	}

	if expected == nil {
		return fieldValue == nil
	}

	if reflect.TypeOf(fieldValue) != reflect.TypeOf(expected) {
		return false
	}

	return reflect.DeepEqual(fieldValue, expected)
}

func findStructFieldValue(value any, fieldName string) (any, bool) {
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return nil, false
	}
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil, false
		}
		v = v.Elem()
	}
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return nil, false
	}
	f := v.FieldByName(fieldName)
	if !f.IsValid() || !f.CanInterface() {
		return nil, false
	}
	return f.Interface(), true
}

func reportTimeoutEvents(t *testing.T, bus *SpyEventBus, eventName string, timeout time.Duration) {
	events := bus.Events()
	logLastCapturedEvents(t, events)
	logLastCapturedEventsOfType(t, events, eventName)
	t.Fatalf("Timeout waiting %v for event of type %s", timeout, eventName)
}

func reportTimeoutEventsByField(t *testing.T, bus *SpyEventBus, eventName, fieldName string, expected any, timeout time.Duration) {
	events := bus.Events()
	logLastCapturedEvents(t, events)
	logLastCapturedEventsOfType(t, events, eventName)
	logLastCapturedEventsOfTypeByField(t, events, eventName, fieldName, expected)
	t.Fatalf("Timeout waiting %v for event of type %s and %s=%v", timeout, eventName, fieldName, expected)
}

func logLastCapturedEvents(t *testing.T, events []event.Event) {
	if len(events) == 0 {
		return
	}

	last := lastEvents(events, spyEventPreviewCount)
	t.Logf("Last %d events captured before timeout (total=%d):", len(last), len(events))
	for i, ev := range last {
		t.Logf("  %d: %T — %+v", i+1, ev, ev)
	}
}

func logLastCapturedEventsOfType(t *testing.T, events []event.Event, eventName string) {
	eventsOfType := filterEventsByType(events, eventName)
	if len(eventsOfType) == 0 {
		t.Logf("No events of type %s were captured before timeout", eventName)
		return
	}

	last := lastEvents(eventsOfType, spyEventPreviewCount)
	t.Logf("Last %d events of type %s captured before timeout (total=%d):", len(last), eventName, len(eventsOfType))
	for i, ev := range last {
		t.Logf("  %d: %T — %+v", i+1, ev, ev)
	}
}

func logLastCapturedEventsOfTypeByField(t *testing.T, events []event.Event, eventName, fieldName string, expected any) {
	eventsOfType := filterEventsByType(events, eventName)
	if len(eventsOfType) == 0 {
		return
	}

	matching := filterEventsByField(eventsOfType, fieldName, expected)
	if len(matching) > 0 {
		last := lastEvents(matching, spyEventPreviewCount)
		t.Logf("Last %d matching events of type %s with %s=%v before timeout (total=%d):", len(last), eventName, fieldName, expected, len(matching))
		for i, ev := range last {
			t.Logf("  %d: %T — %+v", i+1, ev, ev)
		}
		return
	}

	t.Logf("Found %d events of type %s, but none matched %s=%v", len(eventsOfType), eventName, fieldName, expected)

	last := lastEvents(eventsOfType, spyEventPreviewCount)
	t.Logf("Last %d events of type %s with non-matching %s before timeout:", len(last), eventName, fieldName)
	for i, ev := range last {
		fieldValue, ok := findStructFieldValue(ev, fieldName)
		if ok {
			t.Logf("  %d: %T — %s=%v — %+v", i+1, ev, fieldName, fieldValue, ev)
		} else {
			t.Logf("  %d: %T — field %s not found — %+v", i+1, ev, fieldName, ev)
		}
	}
}

func filterEventsByType(events []event.Event, eventName string) []event.Event {
	var filtered []event.Event
	for _, e := range events {
		if utils.GetType(e).String() == eventName {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

func filterEventsByField(events []event.Event, fieldName string, expected any) []event.Event {
	var filtered []event.Event
	for _, e := range events {
		if hasFieldValue(e, fieldName, expected) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

func lastEvents(events []event.Event, count int) []event.Event {
	if len(events) <= count {
		return events
	}
	return events[len(events)-count:]
}
