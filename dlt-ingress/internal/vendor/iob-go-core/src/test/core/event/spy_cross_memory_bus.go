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

type SpyCrossEventBus struct {
	bus    event.CrossBus
	mu     sync.RWMutex
	events []event.CrossEvent
}

func NewSpyCrossEventBus(bus event.CrossBus) *SpyCrossEventBus {
	return &SpyCrossEventBus{
		bus:    bus,
		events: []event.CrossEvent{},
		mu:     sync.RWMutex{},
	}
}

func (s *SpyCrossEventBus) Publish(ctx context.Context, e event.CrossEvent) error {
	s.mu.Lock()
	s.events = append(s.events, e)
	s.mu.Unlock()

	logger.Info(fmt.Sprintf("[SPY-CROSS-BUS] Event %s captured", reflect.TypeOf(e).String()))
	return s.bus.Publish(ctx, e)
}

func (s *SpyCrossEventBus) CleanEvents() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = []event.CrossEvent{}
	logger.Info("[SPY-CROSS-BUS] Event list cleared")

}

func (s *SpyCrossEventBus) Events() []event.CrossEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]event.CrossEvent{}, s.events...)
}

func (s *SpyCrossEventBus) EventsOfType(typeName string) []event.Event {
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

// go
// Deprecated: avoid to use this method. Use ExpectEventProcess or ExpectEventDLT instead
func ExpectCrossEvent[T event.Event](t *testing.T, bus *SpyCrossEventBus, timeout time.Duration) T {
	eventName := utils.GetType((*T)(nil)).String()
	if found, ok := findCapturedCrossEvent[T](bus, eventName); ok {
		return found
	}

	foundChan, stopChan := asyncWaitCrossEvent[T](bus, eventName)

	select {
	case v := <-foundChan:
		close(stopChan)
		return v
	case <-time.After(timeout):
		close(stopChan)
		reportTimeoutCrossEvents(t, bus, eventName, timeout)
		var zero T
		return zero
	}
}

func ExpectCrossEventProcess[T event.Event](t *testing.T, bus *SpyCrossEventBus, processId uuid.UUID, timeout time.Duration) T {
	eventName := utils.GetType((*T)(nil)).String()
	if found, ok := findCapturedCrossEventProcess[T](bus, eventName, processId); ok {
		return found
	}

	foundChan, stopChan := asyncWaitCrossEventProcess[T](bus, eventName, processId)

	select {
	case v := <-foundChan:
		close(stopChan)
		return v
	case <-time.After(timeout):
		close(stopChan)
		reportTimeoutCrossEventsProcess(t, bus, eventName, processId, timeout)
		var zero T
		return zero
	}
}

func ExpectCrossEventDLT[T event.Event](t *testing.T, bus *SpyCrossEventBus, transactionHash string, timeout time.Duration) T {
	eventName := utils.GetType((*T)(nil)).String()
	if found, ok := findCapturedCrossEventDLT[T](bus, eventName, transactionHash); ok {
		return found
	}

	foundChan, stopChan := asyncWaitCrossEventDLT[T](bus, eventName, transactionHash)

	select {
	case v := <-foundChan:
		close(stopChan)
		return v
	case <-time.After(timeout):
		close(stopChan)
		reportTimeoutCrossEventsDLT(t, bus, eventName, transactionHash, timeout)
		var zero T
		return zero
	}
}

func findCapturedCrossEvent[T any](bus *SpyCrossEventBus, eventName string) (T, bool) {
	captured := bus.EventsOfType(eventName)
	for _, e := range captured {
		if typed, ok := e.(T); ok {
			return typed, true
		}
	}
	var zero T
	return zero, false
}

func findCapturedCrossEventProcess[T any](bus *SpyCrossEventBus, eventName string, processId uuid.UUID) (T, bool) {
	captured := bus.EventsOfType(eventName)
	for _, e := range captured {
		if typed, ok := e.(T); ok {
			v := reflect.ValueOf(typed)
			if v.Kind() == reflect.Pointer {
				v = v.Elem()
			}
			if v.IsValid() && v.Kind() == reflect.Struct {
				f := v.FieldByName("ProcessId")
				if f.IsValid() && f.CanInterface() {
					if pid, ok := f.Interface().(uuid.UUID); ok && pid == processId {
						return typed, true
					}
				}
			}
		}
	}
	var zero T
	return zero, false
}

func findCapturedCrossEventDLT[T any](bus *SpyCrossEventBus, eventName, transactionHash string) (T, bool) {
	captured := bus.EventsOfType(eventName)
	for _, e := range captured {
		if typed, ok := e.(T); ok {
			v := reflect.ValueOf(typed)
			if v.Kind() == reflect.Pointer {
				v = v.Elem()
			}
			if v.IsValid() && v.Kind() == reflect.Struct {
				f := v.FieldByName("TransactionHash")
				if f.IsValid() && f.CanInterface() {
					if pid, ok := f.Interface().(string); ok && pid == transactionHash {
						return typed, true
					}
				}
			}
		}
	}
	var zero T
	return zero, false
}

func asyncWaitCrossEvent[T any](bus *SpyCrossEventBus, eventName string) (chan T, chan struct{}) {
	found := make(chan T, 1)
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
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

func asyncWaitCrossEventProcess[T any](bus *SpyCrossEventBus, eventName string, processId uuid.UUID) (chan T, chan struct{}) {
	found := make(chan T, 1)
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				captured := bus.EventsOfType(eventName)
				for _, e := range captured {
					if typed, ok := e.(T); ok {
						v := reflect.ValueOf(typed)
						if v.Kind() == reflect.Pointer {
							v = v.Elem()
						}
						if v.IsValid() && v.Kind() == reflect.Struct {
							f := v.FieldByName("ProcessId")
							if f.IsValid() && f.CanInterface() {
								if pid, ok := f.Interface().(uuid.UUID); ok && pid == processId {
									select {
									case found <- typed:
									default:
									}
									return
								}
							}
						}
					}
				}
			case <-stop:
				return
			}
		}
	}()
	return found, stop
}

func asyncWaitCrossEventDLT[T any](bus *SpyCrossEventBus, eventName, transactionHash string) (chan T, chan struct{}) {
	found := make(chan T, 1)
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				captured := bus.EventsOfType(eventName)
				for _, e := range captured {
					if typed, ok := e.(T); ok {
						v := reflect.ValueOf(typed)
						if v.Kind() == reflect.Pointer {
							v = v.Elem()
						}
						if v.IsValid() && v.Kind() == reflect.Struct {
							f := v.FieldByName("TransactionHash")
							if f.IsValid() && f.CanInterface() {
								if pid, ok := f.Interface().(string); ok && pid == transactionHash {
									select {
									case found <- typed:
									default:
									}
									return
								}
							}
						}
					}
				}
			case <-stop:
				return
			}
		}
	}()
	return found, stop
}

func reportTimeoutCrossEvents(t *testing.T, bus *SpyCrossEventBus, eventName string, timeout time.Duration) {
	const prevEvents = 3
	if len(bus.events) >= prevEvents {
		last := bus.events[len(bus.events)-prevEvents:]
		t.Logf("Last %d events captured before timeout:", len(last))
		for i, ev := range last {
			t.Logf("  %d: %T — %+v", i+1, ev, ev)
		}
	}
	t.Fatalf("Timeout waiting %v for event of type %s", timeout, eventName)
}

func reportTimeoutCrossEventsProcess(t *testing.T, bus *SpyCrossEventBus, eventName string, processId uuid.UUID, timeout time.Duration) {
	const prevEvents = 3
	if len(bus.events) >= prevEvents {
		last := bus.events[len(bus.events)-prevEvents:]
		t.Logf("Last %d events captured before timeout:", len(last))
		for i, ev := range last {
			t.Logf("  %d: %T — %+v", i+1, ev, ev)
		}
	}
	t.Fatalf("Timeout waiting %v for event of type %s and processId %s", timeout, eventName, processId)
}

func reportTimeoutCrossEventsDLT(t *testing.T, bus *SpyCrossEventBus, eventName, transactionHash string, timeout time.Duration) {
	const prevEvents = 3
	if len(bus.events) >= prevEvents {
		last := bus.events[len(bus.events)-prevEvents:]
		t.Logf("Last %d events captured before timeout:", len(last))
		for i, ev := range last {
			t.Logf("  %d: %T — %+v", i+1, ev, ev)
		}
	}
	t.Fatalf("Timeout waiting %v for event of type %s and transactionHash %s", timeout, eventName, transactionHash)
}
