package buildtransactionevents

import "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"

type TransactionBuiltCrossEvent struct {
	event.BaseEvent
	DltAccountId string
	Payload      string
}
