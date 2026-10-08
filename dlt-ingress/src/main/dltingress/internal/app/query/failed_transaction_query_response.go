package query

import "time"

type FailedTransaction struct {
	TxId         string
	NetworkId    string
	CreatedAt    time.Time
	ErrorDetails string
}
