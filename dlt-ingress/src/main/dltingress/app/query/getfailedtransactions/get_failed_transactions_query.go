package getfailedtransactions

import (
	"time"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api/pagination"
)

type Query struct {
	PaginationParams pagination.PaginationParams
}

type Response = pagination.PaginatedQueryResponse[*FailedTransactionQueryResponse]

type FailedTransactionQueryResponse struct {
	TxId         string
	NetworkId    string
	CreatedAt    time.Time
	ErrorDetails string
}
