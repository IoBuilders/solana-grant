//go:build test

package routing

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster/target"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter"
)

// --- stubs ---

type stubEvent struct {
	typ event.Type
}

func (e stubEvent) EventType() event.Type { return e.typ }
func (stubEvent) NodeID() uuid.UUID       { return uuid.Nil }

// stubContractEvent is a hand-written ContractEvent test double.
type stubContractEvent struct {
	eventName string
	status    event.ContractEventStatus
}

func (stubContractEvent) EventType() event.Type                          { return event.TypeContract }
func (stubContractEvent) NodeID() uuid.UUID                              { return uuid.Nil }
func (stubContractEvent) Parameters() []parameter.ContractEventParameter { return nil }
func (e stubContractEvent) EventName() string                            { return e.eventName }
func (e stubContractEvent) Status() event.ContractEventStatus            { return e.status }

// stubSpecification is a Specification test double whose Matches result is
// configured per test.
type stubSpecification struct {
	matches bool
}

func (stubSpecification) Strategy() filter.Strategy          { return filter.Strategy("STUB") }
func (stubSpecification) EventName() string                  { return "stub-event" }
func (s stubSpecification) Matches(event.ContractEvent) bool { return s.matches }

func mustEventFilter(t *testing.T, matches bool) *filter.EventFilter {
	t.Helper()
	f, err := filter.NewEventFilter(
		uuid.New(), mustFilterName(t), uuid.New(), filter.ScopeGlobal, nil, stubSpecification{matches: matches}, nil, nil,
	)
	require.NoError(t, err)
	return f
}

func mustFilterName(t *testing.T) filter.Name {
	t.Helper()
	name, err := filter.NewName("test-filter")
	require.NoError(t, err)
	return name
}

// fakeBroadcasterConfigurationManager is a hand-written test double for
// configurationmanager.BroadcasterConfigurationManager.
type fakeBroadcasterConfigurationManager struct {
	broadcasters []*broadcaster.Broadcaster
	err          error
}

func (m *fakeBroadcasterConfigurationManager) Load(context.Context) ([]*broadcaster.Broadcaster, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.broadcasters, nil
}

// fakeFilterConfigurationManager is a hand-written test double for
// configurationmanager.FilterConfigurationManager.
type fakeFilterConfigurationManager struct {
	filters []filter.Filter
	err     error
}

func (m *fakeFilterConfigurationManager) Load(context.Context) ([]filter.Filter, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.filters, nil
}

func mustBroadcaster(t *testing.T, tg target.Target) *broadcaster.Broadcaster {
	t.Helper()
	b, err := broadcaster.NewBroadcaster(uuid.New(), tg, uuid.New())
	require.NoError(t, err)
	return b
}

func mustDestinations(t *testing.T) []target.Destination {
	t.Helper()
	d, err := target.NewDestination("dest")
	require.NoError(t, err)
	return []target.Destination{d}
}

// --- tests ---

func TestEventRoutingService_Match_BeforeReload_ReturnsEmpty(t *testing.T) {
	s := NewEventRoutingService(&fakeBroadcasterConfigurationManager{}, &fakeFilterConfigurationManager{})

	matched, err := s.Match(context.Background(), stubEvent{typ: event.TypeBlock})

	require.NoError(t, err)
	assert.Empty(t, matched)
}

func TestEventRoutingService_Match_BlockEvent_MatchesBlockAndAllTargets(t *testing.T) {
	destinations := mustDestinations(t)
	blockTarget, err := target.NewBlockTarget(destinations)
	require.NoError(t, err)
	allTarget, err := target.NewAllTarget(destinations)
	require.NoError(t, err)
	txTarget, err := target.NewTransactionTarget(destinations)
	require.NoError(t, err)

	blockBroadcaster := mustBroadcaster(t, blockTarget)
	allBroadcaster := mustBroadcaster(t, allTarget)
	txBroadcaster := mustBroadcaster(t, txTarget)

	s := NewEventRoutingService(&fakeBroadcasterConfigurationManager{
		broadcasters: []*broadcaster.Broadcaster{blockBroadcaster, allBroadcaster, txBroadcaster},
	}, &fakeFilterConfigurationManager{})
	require.NoError(t, s.Reload(context.Background()))

	matched, err := s.Match(context.Background(), stubEvent{typ: event.TypeBlock})

	require.NoError(t, err)
	assert.ElementsMatch(t, []*broadcaster.Broadcaster{blockBroadcaster, allBroadcaster}, matched)
}

func TestEventRoutingService_Match_TransactionEvent_MatchesTransactionAndAllTargets(t *testing.T) {
	destinations := mustDestinations(t)
	txTarget, err := target.NewTransactionTarget(destinations)
	require.NoError(t, err)
	allTarget, err := target.NewAllTarget(destinations)
	require.NoError(t, err)
	blockTarget, err := target.NewBlockTarget(destinations)
	require.NoError(t, err)

	txBroadcaster := mustBroadcaster(t, txTarget)
	allBroadcaster := mustBroadcaster(t, allTarget)
	blockBroadcaster := mustBroadcaster(t, blockTarget)

	s := NewEventRoutingService(&fakeBroadcasterConfigurationManager{
		broadcasters: []*broadcaster.Broadcaster{txBroadcaster, allBroadcaster, blockBroadcaster},
	}, &fakeFilterConfigurationManager{})
	require.NoError(t, s.Reload(context.Background()))

	matched, err := s.Match(context.Background(), stubEvent{typ: event.TypeTransaction})

	require.NoError(t, err)
	assert.ElementsMatch(t, []*broadcaster.Broadcaster{txBroadcaster, allBroadcaster}, matched)
}

func TestEventRoutingService_Match_ContractEvent_MatchesContractEventAndAllTargets(t *testing.T) {
	destinations := mustDestinations(t)
	contractTarget, err := target.NewContractEventTarget(destinations)
	require.NoError(t, err)
	allTarget, err := target.NewAllTarget(destinations)
	require.NoError(t, err)
	filterTarget, err := target.NewFilterTarget(destinations, uuid.New())
	require.NoError(t, err)

	contractBroadcaster := mustBroadcaster(t, contractTarget)
	allBroadcaster := mustBroadcaster(t, allTarget)
	filterBroadcaster := mustBroadcaster(t, filterTarget)

	s := NewEventRoutingService(&fakeBroadcasterConfigurationManager{
		broadcasters: []*broadcaster.Broadcaster{contractBroadcaster, allBroadcaster, filterBroadcaster},
	}, &fakeFilterConfigurationManager{})
	require.NoError(t, s.Reload(context.Background()))

	matched, err := s.Match(context.Background(), stubEvent{typ: event.TypeContract})

	require.NoError(t, err)
	assert.ElementsMatch(t, []*broadcaster.Broadcaster{contractBroadcaster, allBroadcaster}, matched,
		"a FilterTarget broadcaster is only matched when its EventFilter matches — none is configured here")
}

func TestEventRoutingService_Match_ContractEvent_MatchesFilterTargetWhenEventFilterMatches(t *testing.T) {
	destinations := mustDestinations(t)
	matchingFilter := mustEventFilter(t, true)
	filterTarget, err := target.NewFilterTarget(destinations, matchingFilter.ID())
	require.NoError(t, err)
	filterBroadcaster := mustBroadcaster(t, filterTarget)

	nonMatchingFilter := mustEventFilter(t, false)
	otherFilterTarget, err := target.NewFilterTarget(destinations, nonMatchingFilter.ID())
	require.NoError(t, err)
	otherFilterBroadcaster := mustBroadcaster(t, otherFilterTarget)

	s := NewEventRoutingService(
		&fakeBroadcasterConfigurationManager{
			broadcasters: []*broadcaster.Broadcaster{filterBroadcaster, otherFilterBroadcaster},
		},
		&fakeFilterConfigurationManager{
			filters: []filter.Filter{matchingFilter, nonMatchingFilter},
		},
	)
	require.NoError(t, s.Reload(context.Background()))

	matched, err := s.Match(context.Background(), stubContractEvent{eventName: "Transfer", status: event.ContractEventStatusConfirmed})

	require.NoError(t, err)
	assert.Equal(t, []*broadcaster.Broadcaster{filterBroadcaster}, matched)
}

func TestEventRoutingService_Reload_ErrorPropagatesAndKeepsPreviousSnapshot(t *testing.T) {
	destinations := mustDestinations(t)
	blockTarget, err := target.NewBlockTarget(destinations)
	require.NoError(t, err)
	blockBroadcaster := mustBroadcaster(t, blockTarget)

	manager := &fakeBroadcasterConfigurationManager{broadcasters: []*broadcaster.Broadcaster{blockBroadcaster}}
	s := NewEventRoutingService(manager, &fakeFilterConfigurationManager{})
	require.NoError(t, s.Reload(context.Background()))

	manager.err = errors.New("load failed")
	err = s.Reload(context.Background())
	require.Error(t, err)

	matched, err := s.Match(context.Background(), stubEvent{typ: event.TypeBlock})
	require.NoError(t, err)
	assert.Equal(t, []*broadcaster.Broadcaster{blockBroadcaster}, matched, "a failed reload must not clear the existing snapshot")
}

func TestEventRoutingService_Reload_ReplacesPreviousSnapshot(t *testing.T) {
	destinations := mustDestinations(t)
	blockTarget, err := target.NewBlockTarget(destinations)
	require.NoError(t, err)
	oldBroadcaster := mustBroadcaster(t, blockTarget)

	manager := &fakeBroadcasterConfigurationManager{broadcasters: []*broadcaster.Broadcaster{oldBroadcaster}}
	s := NewEventRoutingService(manager, &fakeFilterConfigurationManager{})
	require.NoError(t, s.Reload(context.Background()))

	newBlockTarget, err := target.NewBlockTarget(destinations)
	require.NoError(t, err)
	newBroadcaster := mustBroadcaster(t, newBlockTarget)
	manager.broadcasters = []*broadcaster.Broadcaster{newBroadcaster}
	require.NoError(t, s.Reload(context.Background()))

	matched, err := s.Match(context.Background(), stubEvent{typ: event.TypeBlock})
	require.NoError(t, err)
	assert.Equal(t, []*broadcaster.Broadcaster{newBroadcaster}, matched)
}

// Exercises concurrent Reload/Match; meaningful under `go test -race`.
func TestEventRoutingService_Match_ConcurrentWithReload(t *testing.T) {
	destinations := mustDestinations(t)
	blockTarget, err := target.NewBlockTarget(destinations)
	require.NoError(t, err)
	manager := &fakeBroadcasterConfigurationManager{
		broadcasters: []*broadcaster.Broadcaster{mustBroadcaster(t, blockTarget)},
	}
	s := NewEventRoutingService(manager, &fakeFilterConfigurationManager{})

	var wg sync.WaitGroup
	const iterations = 50
	wg.Add(iterations * 2)
	for i := 0; i < iterations; i++ {
		go func() {
			defer wg.Done()
			_ = s.Reload(context.Background())
		}()
		go func() {
			defer wg.Done()
			_, _ = s.Match(context.Background(), stubEvent{typ: event.TypeBlock})
		}()
	}
	wg.Wait()
}
