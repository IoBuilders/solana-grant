package contractcaller

import (
	"context"
	"fmt"

	"dlt-ingress/src/main/dltingress/internal/infra/evm"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

type ethCallArgs struct {
	To   string `json:"to"`
	Data string `json:"data"`
}

type EvmContractCaller struct {
	evmClientRegistry evm.ClientRegistry
}

func NewEvmContractCaller(ethClientRegistry evm.ClientRegistry) *EvmContractCaller {
	return &EvmContractCaller{evmClientRegistry: ethClientRegistry}
}

func (c *EvmContractCaller) Call(ctx context.Context, request CallRequest) (*CallResponse, error) {
	evmClient, err := c.evmClientRegistry.GetClientForNetworkId(ctx, request.NetworkId)
	if err != nil {
		return nil, err
	}

	blockRef := "latest"
	if request.BlockNumber != nil {
		blockRef = hexutil.EncodeBig(request.BlockNumber)
	}

	var result string
	if err := evmClient.CallRpcMethod(ctx, &result, "eth_call", ethCallArgs{To: request.To, Data: request.Data}, blockRef); err != nil {
		return nil, fmt.Errorf("error calling eth_call on contract %s at block %s: %w", request.To, blockRef, err)
	}

	return &CallResponse{Data: result}, nil
}

var _ Port = (*EvmContractCaller)(nil)
