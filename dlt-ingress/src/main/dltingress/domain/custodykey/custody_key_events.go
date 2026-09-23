package custodykey

import (
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"
)

type KeyCreatedEvent struct {
	event.BaseEvent
	DltAccountId    string
	ExternalId      string
	Dlt             string
	CustodyProvider string
}
