package routing

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/configurationmanager"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster/target"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter"
)

// EventRoutingService is the default Router. It keeps in-memory snapshots of
// Broadcasters and Filters — loaded from a BroadcasterConfigurationManager
// and a FilterConfigurationManager — and matches incoming events against
// them: by Broadcaster Target type for BLOCK/TRANSACTION/CONTRACT_EVENT/ALL
// targets, and, for CONTRACT events, by EventFilter match for FilterTarget
// broadcasters.
type EventRoutingService struct {
	mu                 sync.RWMutex
	broadcasters       []*broadcaster.Broadcaster
	filters            []filter.Filter
	broadcasterManager configurationmanager.BroadcasterConfigurationManager
	filterManager      configurationmanager.FilterConfigurationManager
}

// NewEventRoutingService builds an EventRoutingService backed by
// broadcasterManager and filterManager. Call Reload before the first Match
// to populate the snapshots.
func NewEventRoutingService(
	broadcasterManager configurationmanager.BroadcasterConfigurationManager,
	filterManager configurationmanager.FilterConfigurationManager,
) *EventRoutingService {
	return &EventRoutingService{broadcasterManager: broadcasterManager, filterManager: filterManager}
}

// Reload replaces the in-memory Broadcaster and Filter snapshots with the
// current output of their respective configuration managers.
func (s *EventRoutingService) Reload(ctx context.Context) error {
	broadcasters, err := s.broadcasterManager.Load(ctx)
	if err != nil {
		return err
	}
	filters, err := s.filterManager.Load(ctx)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.broadcasters = broadcasters
	s.filters = filters
	s.mu.Unlock()
	return nil
}

// Match returns every Broadcaster whose Target matches e's type: one whose
// Target.Type() corresponds to e's event type, plus every AllTarget
// broadcaster. For a CONTRACT event, it additionally returns every
// FilterTarget broadcaster whose EventFilter matches e.
func (s *EventRoutingService) Match(_ context.Context, e event.Event) ([]*broadcaster.Broadcaster, error) {
	broadcasters, filters := s.snapshot()
	var matched = make([]*broadcaster.Broadcaster, 0)

	wanted, ok := targetTypeFor(e.EventType())
	if !ok {
		return matched, nil
	}

	for _, b := range broadcasters {
		t := b.Target.Type()
		if t == wanted || t == target.TypeAll {
			matched = append(matched, b)
		}
	}

	if e.EventType() == event.TypeContract {
		if contractEvent, ok := e.(event.ContractEvent); ok {
			matched = append(matched, matchingFilterBroadcasters(broadcasters, filters, contractEvent)...)
		}
	}

	return matched, nil
}

// matchingFilterBroadcasters returns every FilterTarget broadcaster in
// broadcasters whose FilterID identifies an EventFilter in filters that
// matches e.
func matchingFilterBroadcasters(
	broadcasters []*broadcaster.Broadcaster,
	filters []filter.Filter,
	e event.ContractEvent,
) []*broadcaster.Broadcaster {
	matchedFilterIDs := make(map[uuid.UUID]bool)
	for _, f := range filters {
		ef, ok := f.(*filter.EventFilter)
		if !ok {
			continue
		}
		if ef.Matches(e) {
			matchedFilterIDs[ef.ID()] = true
		}
	}
	if len(matchedFilterIDs) == 0 {
		return nil
	}

	var matched []*broadcaster.Broadcaster
	for _, b := range broadcasters {
		ft, ok := b.Target.(*target.FilterTarget)
		if !ok {
			continue
		}
		if matchedFilterIDs[ft.FilterID] {
			matched = append(matched, b)
		}
	}
	return matched
}

func targetTypeFor(t event.Type) (target.Type, bool) {
	switch t {
	case event.TypeBlock:
		return target.TypeBlock, true
	case event.TypeTransaction:
		return target.TypeTransaction, true
	case event.TypeContract:
		return target.TypeContractEvent, true
	case event.TypeSlot:
		return "", false
	default:
		return "", false
	}
}

func (s *EventRoutingService) snapshot() ([]*broadcaster.Broadcaster, []filter.Filter) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	broadcasters := make([]*broadcaster.Broadcaster, len(s.broadcasters))
	copy(broadcasters, s.broadcasters)
	filters := make([]filter.Filter, len(s.filters))
	copy(filters, s.filters)
	return broadcasters, filters
}

var _ Router = (*EventRoutingService)(nil)
