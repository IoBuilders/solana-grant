package queryadapters

import (
	"dlt-ingress/src/main/dltingress/port/query/contractread"
	"dlt-ingress/src/main/dltingress/port/query/gettransaction"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/query"
)

func InitQueryAdapters(queryBus query.Bus) []any {
	return []any{
		gettransactioncross.NewCrossQueryAdapter(queryBus),
		contractreadcross.NewCrossQueryAdapter(queryBus),
	}
}
