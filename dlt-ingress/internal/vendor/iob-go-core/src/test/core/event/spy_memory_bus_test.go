package testevent

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"
)

type MockEvent struct {
	ID string
}

func (m MockEvent) GetId() uuid.UUID {
	id, _ := uuid.NewRandom()
	return id
}

func (m MockEvent) GetCreatedAt() time.Time {
	return time.Now()
}

func (m MockEvent) EventName() string { return "MockEvent" }

func TestCoreSpyEventBus_ConcurrentPublish(t *testing.T) {
	mockBus := &EventBusMock{}
	mockBus.On("Publish", mock.Anything, mock.Anything).Return(nil)
	spy := &SpyEventBus{
		events: []event.Event{},
		mu:     sync.RWMutex{},
		bus:    mockBus,
	}

	const numGoroutines = 50
	const eventsPerGoroutine = 100
	totalExpected := numGoroutines * eventsPerGoroutine

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < eventsPerGoroutine; j++ {
				_ = spy.Publish(context.Background(), MockEvent{})
			}
		}()
	}

	wg.Wait()

	captured := spy.Events()
	assert.Equal(t, totalExpected, len(captured))
}
