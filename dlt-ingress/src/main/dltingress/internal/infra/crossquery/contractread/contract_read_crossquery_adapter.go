package contractreadadapter

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/app/query/contractread"
	"dlt-ingress/src/main/dltingress/port/crossquery/contractread"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/query"
)

type CrossQueryAdapter struct {
	queryBus query.Bus
}

func NewCrossQueryAdapter(queryBus query.Bus) *CrossQueryAdapter {
	return &CrossQueryAdapter{queryBus: queryBus}
}

func (a *CrossQueryAdapter) Execute(ctx context.Context, crossQuery contractreadcross.CrossQuery) (query.CrossResponse, error) {
	result, err := a.queryBus.Dispatch(ctx, contractread.Query{
		SmartContractId:   crossQuery.SmartContractId,
		SmartContractName: crossQuery.SmartContractName,
		MethodName:        crossQuery.MethodName,
		MethodArgs:        crossQuery.MethodArgs,
		NetworkId:         crossQuery.NetworkId,
		BlockNumber:       crossQuery.BlockNumber,
	})
	if err != nil {
		return nil, err
	}

	resp := result.(*contractread.Response)
	return &contractreadcross.Response{
		Result: resp.Result,
	}, nil
}
