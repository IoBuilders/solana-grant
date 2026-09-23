package eventmapper

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	coremapping "gitlab.com/iobuilders/projects/eng/naryo-go/core/app/mapping"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster/target"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

// BlockchainEventSource pairs the event being broadcast with the
// Broadcaster that triggered delivery: some payload fields (e.g. FilterID)
// come from the Broadcaster's Target rather than the event itself.
type BlockchainEventSource struct {
	Event       event.Event
	Broadcaster broadcaster.Broadcaster
}

type EventToJsonMapper struct{}

func NewEventToJsonMapper() *EventToJsonMapper {
	return &EventToJsonMapper{}
}

func (m *EventToJsonMapper) Map(source BlockchainEventSource) ([]byte, error) {
	var payload any
	switch e := source.Event.(type) {
	case event.SolanaContractEvent:
		payload = MapContractEvent(e, filterID(source.Broadcaster))
	case event.SolanaBlockEvent:
		payload = MapBlockEvent(e)
	case event.SolanaTransactionEvent:
		payload = MapTransactionEvent(e)
	default:
		return nil, fmt.Errorf("unsupported event type %T", source.Event)
	}

	return json.Marshal(payload)
}

// filterID returns the FilterID the Broadcaster's Target is scoped to, or
// nil if it isn't a FilterTarget.
func filterID(b broadcaster.Broadcaster) *uuid.UUID {
	ft, ok := b.Target.(*target.FilterTarget)
	if !ok {
		return nil
	}
	return &ft.FilterID
}

var _ coremapping.Mapper[BlockchainEventSource, []byte] = (*EventToJsonMapper)(nil)
