package getfailedtransactions

import (
	"dlt-ingress/src/main/dltingress/internal/app/query"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api/pagination"
)

type Query struct {
	PaginationParams pagination.PaginationParams
}

type Response = pagination.PaginatedQueryResponse[*query.FailedTransaction]
