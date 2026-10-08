package failedtransaction

import "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event"

type SavedEvent struct {
	event.BaseEvent
	TxId         string `json:"txId" required:"true" description:"Identifier of the transaction that failed."`
	NetworkId    string `json:"networkId" required:"true" description:"Network the transaction was submitted to."`
	ErrorDetails string `json:"errorDetails" required:"true" maxLength:"5000" description:"Failure reason as reported by the blockchain context."`
}

type TransitFailedTransactionToRetriedEvent struct {
	event.BaseEvent
	TxId      string `json:"txId" required:"true" description:"Identifier of the failed transaction that the retry superseded."`
	NetworkId string `json:"networkId" required:"true" description:"Network the transaction was submitted to."`
}
