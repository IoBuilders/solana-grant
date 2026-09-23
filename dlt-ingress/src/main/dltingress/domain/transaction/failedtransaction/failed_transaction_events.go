package failedtransaction

import "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"

type SavedEvent struct {
	event.BaseEvent
	TxId         string
	NetworkId    string
	ErrorDetails string
}

type TransitFailedTransactionToRetriedEvent struct {
	event.BaseEvent
	TxId      string
	NetworkId string
}
