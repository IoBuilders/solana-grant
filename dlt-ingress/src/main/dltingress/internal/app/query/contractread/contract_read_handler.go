package contractread

import (
	"context"
	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"dlt-ingress/src/main/dltingress/internal/domain/domainerrors"
	"dlt-ingress/src/main/dltingress/internal/infra/contractcallbuilder"
	"dlt-ingress/src/main/dltingress/internal/infra/contractcaller"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/query"
)

type QueryHandler struct {
	contractCallBuilderRegistry contractcallbuilder.Registry
	contractCallerRegistry      contractcaller.Registry
}

func NewHandler(
	contractCallBuilderRegistry contractcallbuilder.Registry,
	contractCallerRegistry contractcaller.Registry,
) *QueryHandler {
	return &QueryHandler{
		contractCallBuilderRegistry: contractCallBuilderRegistry,
		contractCallerRegistry:      contractCallerRegistry,
	}
}

func (h *QueryHandler) Execute(ctx context.Context, q Query) (query.Response, error) {
	network, err := config.DltIngressConfig.DltIngress.GetDltIngressNetwork(q.NetworkId)
	if err != nil {
		return nil, err
	}

	dlt, err := common.ParseDlt(network.Dlt)
	if err != nil {
		return nil, err
	}

	switch dlt {
	case common.EVM:
		return h.executeEvmCall(ctx, q, dlt)
	default:
		return nil, domainerrors.NewInvalidDltDomainError(network.Dlt)
	}
}

func (h *QueryHandler) executeEvmCall(ctx context.Context, q Query, dlt common.Dlt) (*Response, error) {
	callBuilder, err := h.contractCallBuilderRegistry.GetContractCallBuilder(dlt, q.SmartContractName)
	if err != nil {
		return nil, err
	}

	callData, err := callBuilder.BuildCall(&contractcallbuilder.BuildCallRequest{
		SmartContractId: q.SmartContractId,
		MethodName:      q.MethodName,
		MethodArgs:      q.MethodArgs,
	})
	if err != nil {
		return nil, err
	}

	caller, err := h.contractCallerRegistry.GetContractCaller(dlt)
	if err != nil {
		return nil, err
	}

	callResponse, err := caller.Call(ctx, contractcaller.CallRequest{
		To:          callData.To,
		Data:        callData.Data,
		NetworkId:   q.NetworkId,
		BlockNumber: q.BlockNumber,
	})
	if err != nil {
		return nil, err
	}

	result, err := callBuilder.DecodeResult(q.MethodName, callResponse.Data)
	if err != nil {
		return nil, err
	}

	return &Response{Result: result}, nil
}
