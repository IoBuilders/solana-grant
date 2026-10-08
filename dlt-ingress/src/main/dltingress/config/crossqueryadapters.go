package dltingressconfig

import (
	"dlt-ingress/src/main/dltingress/internal/infra/crossquery/contractread"
	"dlt-ingress/src/main/dltingress/internal/infra/crossquery/gettransaction"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/query"
)

func InitQueryAdapters(queryBus query.Bus) []any {
	return []any{
		gettransactionadapter.NewCrossQueryAdapter(queryBus),
		contractreadadapter.NewCrossQueryAdapter(queryBus),
	}
}
