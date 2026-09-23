package gettransactioncross

import (
	"context"
	"dlt-ingress/src/main/dltingress/app/query/gettransaction"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/query"
)

type CrossQueryAdapter struct {
	queryBus query.Bus
}

func NewCrossQueryAdapter(queryBus query.Bus) *CrossQueryAdapter {
	return &CrossQueryAdapter{queryBus: queryBus}
}

type Response struct {
	gettransaction.Response
}

func (a *CrossQueryAdapter) Execute(ctx context.Context, crossQuery CrossQuery) (query.CrossResponse, error) {

	result, err := a.queryBus.Dispatch(ctx, gettransaction.Query{
		TxId: crossQuery.TxId,
		Dlt:  crossQuery.Dlt,
	})
	if err != nil {
		return nil, err
	}

	resp := result.(gettransaction.Response)
	return &Response{Response: resp}, nil
}
