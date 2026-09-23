package keycreatedlistener

import (
	"context"
	"dlt-ingress/src/main/dltingress/domain/custodykey"
	"dlt-ingress/src/main/dltingress/port/event/custodykey"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"
)

type Listener struct {
	crossEventBus event.CrossBus
}

func NewListener(crossEventBus event.CrossBus) *Listener {
	return &Listener{crossEventBus: crossEventBus}
}

func (l *Listener) Listen(ctx context.Context, evt *custodykey.KeyCreatedEvent) error {
	crossEvt := custodykeyevents.KeyCreatedCrossEvent{
		BaseEvent:       *event.NewBaseEvent(),
		DltAccountId:    evt.DltAccountId,
		ExternalId:      evt.ExternalId,
		Dlt:             evt.Dlt,
		CustodyProvider: evt.CustodyProvider,
	}
	return l.crossEventBus.Publish(ctx, crossEvt)
}
